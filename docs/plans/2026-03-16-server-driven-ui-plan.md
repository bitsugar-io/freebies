# Server-Driven UI Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents
> available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`)
> syntax for tracking.

**Goal:** Backend controls what the mobile app renders — toggle features, reorder content, and add
blocks — without shipping app updates.

**Architecture:** New `GET /api/v1/config` endpoint returns feature flags + ordered screen block
definitions. Mobile app fetches config on launch/foreground, renders blocks from a component
registry. Unknown block types are silently skipped.

**Tech Stack:** Go (chi, sqlc), SQLite/Turso, React Native (Expo), TypeScript

**Spec:** `docs/plans/2026-03-16-server-driven-ui-design.md`

---

## File Map

### Backend (services/api/)

| File | Action | Responsibility |
| ---- | ------ | -------------- |
| `internal/db/migrations/009_server_driven_ui.sql` | Create | Schema + seed data for `feature_flags` and `screen_blocks` tables |
| `db/queries.sql` | Modify | Add `ListFeatureFlags`, `ListAllEnabledScreenBlocks` queries |
| `internal/db/*.go` | Regenerate | sqlc codegen (run `task generate`) |
| `internal/api/handlers/config.go` | Create | `GetConfig` handler |
| `internal/api/handlers/types.go` | Modify | Add config response types |
| `internal/api/server.go` | Modify | Register `GET /api/v1/config` route |
| `internal/api/handlers/config_test.go` | Create | Handler tests |

### Mobile (apps/mobile/)

| File | Action | Responsibility |
| ---- | ------ | -------------- |
| `src/api/client.ts` | Modify | Add `AppConfig`, `ScreenBlock` types and `getConfig()` method |
| `src/context/AppConfigContext.tsx` | Create | Config provider, fetch on mount + foreground, fallback to `DEFAULT_CONFIG` |
| `src/components/blocks/BlockRenderer.tsx` | Create | Maps block type → component, renders block list for a screen |
| `src/components/blocks/BannerBlock.tsx` | Create | Dismissible banner with text/color from config |
| `src/components/blocks/ActiveDealsBlock.tsx` | Create | Wraps existing active deals list |
| `src/components/blocks/LeagueFilterBlock.tsx` | Create | Wraps existing league filter bar |
| `src/components/blocks/EventListBlock.tsx` | Create | Wraps existing grouped event list |
| `src/components/blocks/PromoCardBlock.tsx` | Create | CTA card with title, subtitle, link, colors |
| `src/components/blocks/UserStatsBlock.tsx` | Create | Wraps existing user stats display |
| `src/components/blocks/SubscriptionListBlock.tsx` | Create | Wraps existing subscription list |
| `src/components/blocks/SettingsBlock.tsx` | Create | Wraps existing settings section |
| `src/screens/DealsScreen.tsx` | Modify | Replace hardcoded layout with `BlockRenderer` |
| `src/screens/DiscoverScreen.tsx` | Modify | Replace hardcoded layout with `BlockRenderer` |
| `src/screens/ProfileScreen.tsx` | Modify | Replace hardcoded layout with `BlockRenderer` |
| `App.tsx` | Modify | Wrap app in `AppConfigProvider` |

---

## Chunk 1: Backend — Database & Queries

### Task 1: Create migration

**Files:**
- Create: `services/api/internal/db/migrations/009_server_driven_ui.sql`

- [ ] **Step 1: Create the migration file**

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS feature_flags (
    key TEXT PRIMARY KEY,
    enabled INTEGER NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS screen_blocks (
    id TEXT PRIMARY KEY,
    screen TEXT NOT NULL,
    type TEXT NOT NULL,
    key TEXT NOT NULL UNIQUE,
    position INTEGER NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 1,
    config TEXT NOT NULL DEFAULT '{}',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_screen_blocks_screen ON screen_blocks(screen);

-- Seed feature flags
INSERT INTO feature_flags (key, enabled) VALUES
    ('enable_mlb', 1),
    ('enable_nba', 1),
    ('enable_nfl', 1),
    ('enable_nhl', 0),
    ('show_affiliate_links', 0),
    ('maintenance_mode', 0);

-- Seed screen blocks (matches current hardcoded layout)
INSERT INTO screen_blocks (id, screen, type, key, position, config) VALUES
    ('blk_deals_active', 'deals', 'active_deals', 'active-deals-list', 1,
     '{"layout":"list","emptyTitle":"No Active Deals","emptySubtitle":"Deals appear here when your teams trigger offers"}');

INSERT INTO screen_blocks (id, screen, type, key, position, config) VALUES
    ('blk_discover_filter', 'discover', 'league_filter', 'league-filter-bar', 1, '{}'),
    ('blk_discover_events', 'discover', 'event_list', 'event-list', 2, '{"groupBy":"team"}');

INSERT INTO screen_blocks (id, screen, type, key, position, config) VALUES
    ('blk_profile_stats', 'profile', 'user_stats', 'user-stats', 1, '{}'),
    ('blk_profile_subs', 'profile', 'subscription_list', 'subscriptions', 2, '{}'),
    ('blk_profile_settings', 'profile', 'settings', 'settings', 3, '{"showThemeToggle":true}');

-- +goose Down
DROP INDEX IF EXISTS idx_screen_blocks_screen;
DROP TABLE IF EXISTS screen_blocks;
DROP TABLE IF EXISTS feature_flags;
```

- [ ] **Step 2: Verify migration runs**

Run: `task clean && task api:serve`

Expected: Server starts, logs show migration 009 applied.

- [ ] **Step 3: Commit**

```bash
git add services/api/internal/db/migrations/009_server_driven_ui.sql
git commit -m "feat(db): Add feature_flags and screen_blocks tables"
```

### Task 2: Add sqlc queries and regenerate

**Files:**
- Modify: `services/api/db/queries.sql`
- Regenerate: `services/api/internal/db/` (sqlc output)

- [ ] **Step 1: Add queries to queries.sql**

Append to the end of `services/api/db/queries.sql`:

```sql
-- name: ListFeatureFlags :many
SELECT * FROM feature_flags ORDER BY key;

-- name: ListAllEnabledScreenBlocks :many
SELECT * FROM screen_blocks
WHERE enabled = 1
ORDER BY screen, position ASC;
```

- [ ] **Step 2: Regenerate sqlc code**

Run: `task generate`

Expected: No errors. New methods `ListFeatureFlags` and `ListAllEnabledScreenBlocks` appear in
generated code.

- [ ] **Step 3: Verify it compiles**

Run: `cd services/api && go build ./...`

Expected: Clean build, no errors.

- [ ] **Step 4: Commit**

```bash
git add services/api/db/queries.sql services/api/internal/db/
git commit -m "feat(db): Add config queries for feature flags and screen blocks"
```

---

## Chunk 2: Backend — Config Handler & Route

### Task 3: Add config response types

**Files:**
- Modify: `services/api/internal/api/handlers/types.go`

- [ ] **Step 1: Add config types to types.go**

Append to the end of `services/api/internal/api/handlers/types.go`:

```go
// ConfigResponse is the top-level config response
type ConfigResponse struct {
	Features map[string]bool            `json:"features"`
	Screens  map[string][]ScreenBlock   `json:"screens"`
}

// ScreenBlock represents a UI block in a screen layout
type ScreenBlock struct {
	Type   string                 `json:"type"`
	Key    string                 `json:"key"`
	Config map[string]interface{} `json:"config"`
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd services/api && go build ./...`

Expected: Clean build.

- [ ] **Step 3: Commit**

```bash
git add services/api/internal/api/handlers/types.go
git commit -m "feat(api): Add config response types"
```

### Task 4: Create GetConfig handler

**Files:**
- Create: `services/api/internal/api/handlers/config.go`

- [ ] **Step 1: Write the handler**

```go
package handlers

import (
	"encoding/json"
	"net/http"
)

// GetConfig returns feature flags and screen block layouts.
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch feature flags
	flags, err := h.queries.ListFeatureFlags(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load config")
		return
	}

	features := make(map[string]bool, len(flags))
	for _, f := range flags {
		features[f.Key] = f.Enabled == 1
	}

	// Fetch enabled screen blocks
	blocks, err := h.queries.ListAllEnabledScreenBlocks(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load config")
		return
	}

	screens := make(map[string][]ScreenBlock)
	for _, b := range blocks {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(b.Config), &config); err != nil {
			config = make(map[string]interface{})
		}

		screens[b.Screen] = append(screens[b.Screen], ScreenBlock{
			Type:   b.Type,
			Key:    b.Key,
			Config: config,
		})
	}

	w.Header().Set("Cache-Control", "public, max-age=60")
	respondJSON(w, http.StatusOK, ConfigResponse{
		Features: features,
		Screens:  screens,
	})
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd services/api && go build ./...`

Expected: Clean build. Note — the `Enabled` and `Config` field names come from sqlc codegen. If
the generated field names differ (e.g., `Enabled` is `int64` not `int`), adjust the comparison
accordingly. Check `services/api/internal/db/models.go` for exact types.

- [ ] **Step 3: Commit**

```bash
git add services/api/internal/api/handlers/config.go
git commit -m "feat(api): Add GetConfig handler"
```

### Task 5: Register the route

**Files:**
- Modify: `services/api/internal/api/server.go`

- [ ] **Step 1: Add config route**

In `server.go`, inside the `s.router.Route("/api/v1", ...)` block, add with the other public routes
(near `ListLeagues`, `ListEvents`):

```go
r.Get("/config", h.GetConfig)
```

- [ ] **Step 2: Verify it compiles**

Run: `cd services/api && go build ./...`

Expected: Clean build.

- [ ] **Step 3: Commit**

```bash
git add services/api/internal/api/server.go
git commit -m "feat(api): Register GET /api/v1/config route"
```

### Task 6: Write handler test

**Files:**
- Create: `services/api/internal/api/handlers/config_test.go`

- [ ] **Step 1: Write the test**

There are no existing handler tests in this project. This test uses a real in-memory SQLite
database with goose migrations (same as production). The `db.Open` function uses `go-sqlite3` for
local paths and the `db.Migrate` function runs all embedded goose migrations.

```go
package handlers_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/retr0h/freebie/services/api/internal/api/handlers"
	"github.com/retr0h/freebie/services/api/internal/db"
)

// setupTestHandler creates a handler backed by an in-memory SQLite database
// with all migrations applied (including seed data).
func setupTestHandler(t *testing.T) *handlers.Handler {
	t.Helper()

	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	if err := db.Migrate(database); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return handlers.New(database, logger)
}

func TestGetConfig(t *testing.T) {
	h := setupTestHandler(t)

	t.Run("returns features and screens from seed data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
		w := httptest.NewRecorder()

		h.GetConfig(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var config handlers.ConfigResponse
		if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// Verify feature flags from seed data
		if len(config.Features) != 6 {
			t.Fatalf("expected 6 feature flags, got %d", len(config.Features))
		}
		if config.Features["enable_mlb"] != true {
			t.Fatal("expected enable_mlb to be true")
		}
		if config.Features["enable_nhl"] != false {
			t.Fatal("expected enable_nhl to be false")
		}

		// Verify screens from seed data
		if len(config.Screens) != 3 {
			t.Fatalf("expected 3 screens, got %d", len(config.Screens))
		}
		if len(config.Screens["deals"]) != 1 {
			t.Fatalf("expected 1 block in deals, got %d", len(config.Screens["deals"]))
		}
		if config.Screens["deals"][0].Type != "active_deals" {
			t.Fatalf("expected active_deals block, got %s", config.Screens["deals"][0].Type)
		}
		if len(config.Screens["discover"]) != 2 {
			t.Fatalf("expected 2 blocks in discover, got %d", len(config.Screens["discover"]))
		}
		if len(config.Screens["profile"]) != 3 {
			t.Fatalf("expected 3 blocks in profile, got %d", len(config.Screens["profile"]))
		}

		// Verify cache header
		cc := resp.Header.Get("Cache-Control")
		if cc != "public, max-age=60" {
			t.Fatalf("expected Cache-Control 'public, max-age=60', got %q", cc)
		}

		// Verify config blob is parsed (not a raw string)
		dealsConfig := config.Screens["deals"][0].Config
		if dealsConfig["layout"] != "list" {
			t.Fatalf("expected deals config layout=list, got %v", dealsConfig["layout"])
		}
	})
}
```

- [ ] **Step 2: Run tests**

Run: `task api:test`

Expected: All tests pass including the new config test.

- [ ] **Step 3: Commit**

```bash
git add services/api/internal/api/handlers/config_test.go
git commit -m "test(api): Add GetConfig handler test"
```

### Task 7: Manual smoke test

- [ ] **Step 1: Start the server**

Run: `task clean && task api:serve`

- [ ] **Step 2: Test the endpoint**

Run: `curl -s http://localhost:8080/api/v1/config | jq .`

Expected: JSON response with `features` object (6 flags) and `screens` object (3 screens with
blocks matching seed data). Verify `Cache-Control: public, max-age=60` header is present.

---

## Chunk 3: Mobile — Config Context & API Client

### Task 8: Add config types and API method

**Files:**
- Modify: `apps/mobile/src/api/client.ts`

- [ ] **Step 1: Add types**

Add these types near the other interface definitions in `client.ts`:

```typescript
export interface ScreenBlock {
  type: string;
  key: string;
  config: Record<string, any>;
}

export interface AppConfig {
  features: Record<string, boolean>;
  screens: Record<string, ScreenBlock[]>;
}
```

- [ ] **Step 2: Add API method**

Add this method to the `ApiClient` class (alongside existing methods like `listLeagues`). This is a
public endpoint — no auth required:

```typescript
  async getConfig(): Promise<AppConfig> {
    return this.request<AppConfig>('/config');
  }
```

- [ ] **Step 3: Commit**

```bash
git add apps/mobile/src/api/client.ts
git commit -m "feat(mobile): Add config types and getConfig API method"
```

### Task 9: Create AppConfigContext

**Files:**
- Create: `apps/mobile/src/context/AppConfigContext.tsx`

- [ ] **Step 1: Write the context provider**

Follow the same patterns used in `AppDataContext.tsx` — `useState`, `useEffect`, `AppState`
listener. Include a `DEFAULT_CONFIG` fallback.

```typescript
import React, { createContext, useContext, useState, useEffect, useCallback, useRef } from 'react';
import { AppState, AppStateStatus } from 'react-native';
import { api, AppConfig } from '../api/client';

const DEFAULT_CONFIG: AppConfig = {
  features: {
    enable_mlb: true,
    enable_nba: true,
    enable_nfl: true,
    enable_nhl: false,
    show_affiliate_links: false,
    maintenance_mode: false,
  },
  screens: {
    deals: [
      { type: 'active_deals', key: 'active-deals-list', config: { layout: 'list', emptyTitle: 'No Active Deals', emptySubtitle: 'Deals appear here when your teams trigger offers' } },
    ],
    discover: [
      { type: 'league_filter', key: 'league-filter-bar', config: {} },
      { type: 'event_list', key: 'event-list', config: { groupBy: 'team' } },
    ],
    profile: [
      { type: 'user_stats', key: 'user-stats', config: {} },
      { type: 'subscription_list', key: 'subscriptions', config: {} },
      { type: 'settings', key: 'settings', config: { showThemeToggle: true } },
    ],
  },
};

interface AppConfigContextType {
  config: AppConfig;
  isLoading: boolean;
  refreshConfig: () => Promise<void>;
}

const AppConfigContext = createContext<AppConfigContextType>({
  config: DEFAULT_CONFIG,
  isLoading: true,
  refreshConfig: async () => {},
});

export function AppConfigProvider({ children }: { children: React.ReactNode }) {
  const [config, setConfig] = useState<AppConfig>(DEFAULT_CONFIG);
  const [isLoading, setIsLoading] = useState(true);
  const appState = useRef(AppState.currentState);

  const refreshConfig = useCallback(async () => {
    try {
      const result = await api.getConfig();
      setConfig(result);
    } catch (err) {
      console.warn('Failed to fetch config, using defaults:', err);
      // Keep current config (or DEFAULT_CONFIG on first load)
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Fetch on mount
  useEffect(() => {
    refreshConfig();
  }, [refreshConfig]);

  // Refresh on foreground
  useEffect(() => {
    const sub = AppState.addEventListener('change', (nextState: AppStateStatus) => {
      if (appState.current.match(/inactive|background/) && nextState === 'active') {
        refreshConfig();
      }
      appState.current = nextState;
    });
    return () => sub.remove();
  }, [refreshConfig]);

  return (
    <AppConfigContext.Provider value={{ config, isLoading, refreshConfig }}>
      {children}
    </AppConfigContext.Provider>
  );
}

export function useAppConfig() {
  return useContext(AppConfigContext);
}

export { DEFAULT_CONFIG };
```

- [ ] **Step 2: Wrap app in provider**

In `App.tsx`, import `AppConfigProvider` and wrap it around the existing providers. It should go
**outside** `AppDataProvider` since config doesn't depend on user data:

```typescript
import { AppConfigProvider } from './src/context/AppConfigContext';
```

Then in the JSX, wrap:

```tsx
<AppConfigProvider>
  {/* existing AppDataProvider, NavigationContainer, etc. */}
</AppConfigProvider>
```

- [ ] **Step 3: Verify app still runs**

Run: `task mobile:serve`, press `i` for iOS simulator.

Expected: App loads normally. Console may show config fetch (or warning if backend isn't running
with new migration). No crashes.

- [ ] **Step 4: Commit**

```bash
git add apps/mobile/src/context/AppConfigContext.tsx apps/mobile/App.tsx
git commit -m "feat(mobile): Add AppConfigContext with foreground refresh"
```

---

## Chunk 4: Mobile — Block Renderer & Block Components

### Task 10: Create BlockRenderer

**Files:**
- Create: `apps/mobile/src/components/blocks/BlockRenderer.tsx`

- [ ] **Step 1: Write the renderer**

```typescript
import React from 'react';
import { ScrollView, RefreshControl } from 'react-native';
import { ScreenBlock } from '../../api/client';
import { useAppConfig } from '../../context/AppConfigContext';
import { BannerBlock } from './BannerBlock';
import { ActiveDealsBlock } from './ActiveDealsBlock';
import { LeagueFilterBlock } from './LeagueFilterBlock';
import { EventListBlock } from './EventListBlock';
import { PromoCardBlock } from './PromoCardBlock';
import { UserStatsBlock } from './UserStatsBlock';
import { SubscriptionListBlock } from './SubscriptionListBlock';
import { SettingsBlock } from './SettingsBlock';

export interface BlockProps {
  config: Record<string, any>;
}

const BLOCK_REGISTRY: Record<string, React.ComponentType<BlockProps>> = {
  banner: BannerBlock,
  active_deals: ActiveDealsBlock,
  league_filter: LeagueFilterBlock,
  event_list: EventListBlock,
  promo_card: PromoCardBlock,
  user_stats: UserStatsBlock,
  subscription_list: SubscriptionListBlock,
  settings: SettingsBlock,
};

interface BlockRendererProps {
  screenId: string;
  refreshing?: boolean;
  onRefresh?: () => void;
}

export function BlockRenderer({ screenId, refreshing, onRefresh }: BlockRendererProps) {
  const { config } = useAppConfig();
  const blocks = config?.screens[screenId] ?? [];

  return (
    <ScrollView
      refreshControl={
        onRefresh ? (
          <RefreshControl refreshing={refreshing ?? false} onRefresh={onRefresh} />
        ) : undefined
      }
    >
      {blocks.map((block) => {
        const Component = BLOCK_REGISTRY[block.type];
        if (!Component) return null;
        return <Component key={block.key} config={block.config} />;
      })}
    </ScrollView>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add apps/mobile/src/components/blocks/BlockRenderer.tsx
git commit -m "feat(mobile): Add BlockRenderer component"
```

### Task 11: Create block components

**Files:**
- Create: `apps/mobile/src/components/blocks/BannerBlock.tsx`
- Create: `apps/mobile/src/components/blocks/ActiveDealsBlock.tsx`
- Create: `apps/mobile/src/components/blocks/LeagueFilterBlock.tsx`
- Create: `apps/mobile/src/components/blocks/EventListBlock.tsx`
- Create: `apps/mobile/src/components/blocks/PromoCardBlock.tsx`
- Create: `apps/mobile/src/components/blocks/UserStatsBlock.tsx`
- Create: `apps/mobile/src/components/blocks/SubscriptionListBlock.tsx`
- Create: `apps/mobile/src/components/blocks/SettingsBlock.tsx`

Each block component wraps existing screen logic into a standalone component that receives `config`
as a prop. The goal is to **extract** the existing rendering code from each screen, not rewrite it.

- [ ] **Step 1: Read existing screen files**

Read `DealsScreen.tsx`, `DiscoverScreen.tsx`, and `ProfileScreen.tsx` fully to understand the
current rendering logic. Each block component should extract the relevant section from its screen.

- [ ] **Step 2: Create BannerBlock**

This is a new component (no existing code to extract):

```typescript
import React, { useState } from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { BlockProps } from './BlockRenderer';
import { useTheme } from '../../hooks/useTheme';

export function BannerBlock({ config }: BlockProps) {
  const [dismissed, setDismissed] = useState(false);
  const { colors } = useTheme();

  if (dismissed) return null;

  const text = config.text as string;
  if (!text) return null;

  const backgroundColor = (config.backgroundColor as string) ?? colors.card;
  const textColor = (config.textColor as string) ?? colors.text;
  const dismissible = config.dismissible !== false;

  return (
    <View style={[styles.container, { backgroundColor }]}>
      <Text style={[styles.text, { color: textColor }]}>{text}</Text>
      {dismissible && (
        <TouchableOpacity onPress={() => setDismissed(true)} style={styles.dismiss}>
          <Text style={{ color: textColor, opacity: 0.7 }}>✕</Text>
        </TouchableOpacity>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
    marginHorizontal: 16,
    marginTop: 16,
    borderRadius: 12,
  },
  text: {
    flex: 1,
    fontSize: 15,
    fontWeight: '600',
  },
  dismiss: {
    paddingLeft: 12,
  },
});
```

- [ ] **Step 3: Create ActiveDealsBlock**

Extracted from `DealsScreen.tsx` — the FlatList of ActiveDealCard items with empty state. The
modal handling and deal press logic stay in DealsScreen (the block just renders the list).

```typescript
// src/components/blocks/ActiveDealsBlock.tsx
import React from 'react';
import { View, Text, FlatList, StyleSheet } from 'react-native';
import { BlockProps } from './BlockRenderer';
import { useAppData } from '../../context/AppDataContext';
import { useTheme } from '../../hooks/useTheme';
import { ActiveDealCard } from '../ActiveDealCard';

export function ActiveDealsBlock({ config }: BlockProps) {
  const { theme } = useTheme();
  const { colors } = theme;
  const { undismissedDeals, dismissDeal } = useAppData();

  const emptyTitle = (config.emptyTitle as string) ?? 'No Active Deals';
  const emptySubtitle = (config.emptySubtitle as string) ??
    'Subscribe to events in the Discover tab to get notified when deals drop!';

  return (
    <FlatList
      scrollEnabled={false}
      data={undismissedDeals}
      renderItem={({ item }) => (
        <ActiveDealCard
          deal={item}
          onPress={() => {}}  // Parent screen handles modal via onDealPress prop
          onDismiss={(type) => dismissDeal(item.id, type)}
        />
      )}
      keyExtractor={(item) => item.id}
      contentContainerStyle={styles.listContent}
      ListEmptyComponent={
        <View style={styles.emptyContainer}>
          <Text style={styles.emptyIcon}>🎁</Text>
          <Text style={[styles.emptyTitle, { color: colors.text }]}>{emptyTitle}</Text>
          <Text style={[styles.emptyText, { color: colors.textMuted }]}>{emptySubtitle}</Text>
        </View>
      }
    />
  );
}

const styles = StyleSheet.create({
  listContent: { padding: 16, paddingBottom: 100 },
  emptyContainer: { padding: 32, alignItems: 'center', marginTop: 40 },
  emptyIcon: { fontSize: 48, marginBottom: 16 },
  emptyTitle: { fontSize: 20, fontWeight: '600', marginBottom: 8 },
  emptyText: { fontSize: 16, textAlign: 'center', lineHeight: 22 },
});
```

Note: The `onPress` handler for deal taps needs to be wired up. The simplest approach is to have
`DealsScreen` pass a callback through context or keep the modal logic in the screen and have the
block emit events upward. For v1, keep the `DealModal` in `DealsScreen` and add an `onDealPress`
prop to `ActiveDealsBlock`:

```typescript
// Update BlockProps to support optional callbacks
export interface BlockProps {
  config: Record<string, any>;
  onDealPress?: (deal: ActiveDeal) => void;
}
```

Then `BlockRenderer` passes screen-level callbacks through to blocks. Alternatively, use a context
for modal state — follow whichever pattern feels cleaner when implementing.

- [ ] **Step 4: Create LeagueFilterBlock**

Extracted from `DiscoverScreen.tsx` — the horizontal ScrollView of league pills.

```typescript
// src/components/blocks/LeagueFilterBlock.tsx
import React, { useMemo } from 'react';
import { View, Text, ScrollView, TouchableOpacity, StyleSheet } from 'react-native';
import { BlockProps } from './BlockRenderer';
import { useAppData } from '../../context/AppDataContext';
import { useTheme } from '../../hooks/useTheme';

interface LeagueFilterProps extends BlockProps {
  selectedLeague: string;
  onSelectLeague: (id: string) => void;
}

export function LeagueFilterBlock({ config, selectedLeague, onSelectLeague }: LeagueFilterProps) {
  const { theme } = useTheme();
  const { colors } = theme;
  const { leagues } = useAppData();

  const leagueOptions = useMemo(() => {
    const all = { id: 'all', name: 'All', icon: '🌟', displayOrder: 0 };
    return [all, ...leagues];
  }, [leagues]);

  return (
    <View style={[styles.container, { backgroundColor: colors.surface }]}>
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={styles.scroll}
      >
        {leagueOptions.map(league => {
          const isActive = selectedLeague === league.id;
          return (
            <TouchableOpacity
              key={league.id}
              style={[styles.pill, { backgroundColor: isActive ? colors.accent : colors.surfaceSecondary }]}
              onPress={() => onSelectLeague(league.id)}
            >
              <Text style={styles.icon}>{league.icon}</Text>
              <Text style={[styles.label, { color: isActive ? '#fff' : colors.text }]}>
                {league.name}
              </Text>
            </TouchableOpacity>
          );
        })}
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { paddingVertical: 12 },
  scroll: { paddingHorizontal: 12 },
  pill: {
    flexDirection: 'row', alignItems: 'center',
    paddingHorizontal: 16, paddingVertical: 10,
    borderRadius: 20, marginRight: 8,
  },
  icon: { fontSize: 16, marginRight: 6 },
  label: { fontSize: 14, fontWeight: '600' },
});
```

Note: `LeagueFilterBlock` needs `selectedLeague` and `onSelectLeague` as props — these come from
the parent `DiscoverScreen` which owns the filter state. This means `BlockRenderer` needs to pass
screen-specific props to certain block types, or the filter state moves into a shared context. For
v1, the simplest approach is to keep the filter state in `DiscoverScreen` and pass it through the
block renderer. See Task 12 for how this is wired.

- [ ] **Step 5: Create EventListBlock**

Extracted from `DiscoverScreen.tsx` — the team-grouped FlatList with expandable cards.

```typescript
// src/components/blocks/EventListBlock.tsx
import React, { useState, useMemo } from 'react';
import { View, Text, FlatList, TouchableOpacity, StyleSheet } from 'react-native';
import { BlockProps } from './BlockRenderer';
import { useAppData } from '../../context/AppDataContext';
import { useTheme } from '../../hooks/useTheme';
import { Event } from '../../api/client';

const DEFAULT_TEAM_COLOR = '#666666';

interface TeamGroup {
  teamId: string;
  teamName: string;
  league: string;
  color: string;
  events: Event[];
  isFollowing: boolean;
  followedCount: number;
}

interface EventListProps extends BlockProps {
  selectedLeague?: string;
  onEventPress?: (event: Event) => void;
}

export function EventListBlock({ config, selectedLeague = 'all', onEventPress }: EventListProps) {
  const { theme } = useTheme();
  const { colors } = theme;
  const { events, isSubscribed, toggleSubscription } = useAppData();
  const [expandedTeam, setExpandedTeam] = useState<string | null>(null);

  const teamGroups = useMemo(() => {
    const groups: Record<string, TeamGroup> = {};
    events.forEach(event => {
      if (!groups[event.teamId]) {
        groups[event.teamId] = {
          teamId: event.teamId, teamName: event.teamName,
          league: event.league.toLowerCase(), color: event.teamColor || DEFAULT_TEAM_COLOR,
          events: [], isFollowing: false, followedCount: 0,
        };
      }
      groups[event.teamId].events.push(event);
      if (isSubscribed(event.id)) groups[event.teamId].followedCount++;
    });
    Object.values(groups).forEach(g => {
      g.isFollowing = g.events.length > 0 && g.events.every(e => isSubscribed(e.id));
    });
    return Object.values(groups);
  }, [events, isSubscribed]);

  const filteredTeams = useMemo(() => {
    if (selectedLeague === 'all') return teamGroups;
    return teamGroups.filter(t => t.league === selectedLeague);
  }, [teamGroups, selectedLeague]);

  const handleFollowTeam = async (team: TeamGroup) => {
    for (const event of team.events) {
      if (team.isFollowing) {
        if (isSubscribed(event.id)) await toggleSubscription(event.id);
      } else {
        if (!isSubscribed(event.id)) await toggleSubscription(event.id);
      }
    }
  };

  const renderTeamCard = ({ item: team }: { item: TeamGroup }) => {
    const isExpanded = expandedTeam === team.teamId;
    return (
      <View style={[styles.teamCard, { backgroundColor: colors.surface }]}>
        <TouchableOpacity
          style={styles.teamHeader}
          onPress={() => setExpandedTeam(isExpanded ? null : team.teamId)}
          activeOpacity={0.7}
        >
          <View style={[styles.teamBadge, { backgroundColor: team.color }]}>
            <Text style={styles.teamBadgeText}>{team.teamId}</Text>
          </View>
          <View style={styles.teamInfo}>
            <Text style={[styles.teamName, { color: colors.text }]}>{team.teamName}</Text>
            <Text style={[styles.teamOffers, { color: colors.textMuted }]}>
              {team.events.length} offer{team.events.length !== 1 ? 's' : ''} available
              {team.followedCount > 0 && ` • ${team.followedCount} following`}
            </Text>
          </View>
          <TouchableOpacity
            style={[styles.followButton, {
              backgroundColor: team.isFollowing ? colors.surfaceSecondary : colors.accent
            }]}
            onPress={() => handleFollowTeam(team)}
          >
            <Text style={[styles.followButtonText, {
              color: team.isFollowing ? colors.text : '#fff'
            }]}>
              {team.isFollowing ? 'Following' : 'Follow'}
            </Text>
          </TouchableOpacity>
        </TouchableOpacity>

        {isExpanded && (
          <View style={[styles.eventsList, { borderTopColor: colors.border }]}>
            {team.events.map(event => (
              <TouchableOpacity
                key={event.id}
                style={[styles.eventRow, { borderBottomColor: colors.border }]}
                onPress={() => onEventPress?.(event)}
              >
                <View style={styles.eventInfo}>
                  <Text style={[styles.eventName, { color: colors.text }]}>{event.offerName}</Text>
                  <Text style={[styles.eventPartner, { color: colors.textMuted }]}>
                    {event.partnerName} • {event.triggerCondition}
                  </Text>
                </View>
                <TouchableOpacity
                  style={[styles.subscribeButton, {
                    backgroundColor: isSubscribed(event.id) ? colors.success : colors.surfaceSecondary
                  }]}
                  onPress={() => toggleSubscription(event.id)}
                >
                  <Text style={[styles.subscribeButtonText, {
                    color: isSubscribed(event.id) ? '#fff' : colors.textMuted
                  }]}>
                    {isSubscribed(event.id) ? '✓' : '+'}
                  </Text>
                </TouchableOpacity>
              </TouchableOpacity>
            ))}
          </View>
        )}

        <TouchableOpacity
          style={styles.expandHint}
          onPress={() => setExpandedTeam(isExpanded ? null : team.teamId)}
        >
          <Text style={[styles.expandHintText, { color: colors.textMuted }]}>
            {isExpanded ? '▲ Collapse' : '▼ See offers'}
          </Text>
        </TouchableOpacity>
      </View>
    );
  };

  return (
    <FlatList
      scrollEnabled={false}
      data={filteredTeams}
      renderItem={renderTeamCard}
      keyExtractor={(item) => item.teamId}
      contentContainerStyle={styles.listContent}
      ListEmptyComponent={
        <View style={styles.emptyContainer}>
          <Text style={styles.emptyIcon}>🏟️</Text>
          <Text style={[styles.emptyTitle, { color: colors.text }]}>No Teams Found</Text>
          <Text style={[styles.emptyText, { color: colors.textMuted }]}>
            No teams available in this league yet.
          </Text>
        </View>
      }
    />
  );
}

const styles = StyleSheet.create({
  listContent: { padding: 16, paddingBottom: 100 },
  teamCard: { borderRadius: 12, marginBottom: 12, overflow: 'hidden' },
  teamHeader: { flexDirection: 'row', alignItems: 'center', padding: 16 },
  teamBadge: { width: 48, height: 48, borderRadius: 24, alignItems: 'center', justifyContent: 'center' },
  teamBadgeText: { color: '#fff', fontWeight: 'bold', fontSize: 14 },
  teamInfo: { flex: 1, marginLeft: 12 },
  teamName: { fontSize: 18, fontWeight: '600' },
  teamOffers: { fontSize: 13, marginTop: 2 },
  followButton: { paddingHorizontal: 16, paddingVertical: 8, borderRadius: 16 },
  followButtonText: { fontSize: 14, fontWeight: '600' },
  eventsList: { borderTopWidth: 1 },
  eventRow: { flexDirection: 'row', alignItems: 'center', paddingHorizontal: 16, paddingVertical: 12, borderBottomWidth: StyleSheet.hairlineWidth },
  eventInfo: { flex: 1 },
  eventName: { fontSize: 15, fontWeight: '500' },
  eventPartner: { fontSize: 12, marginTop: 2 },
  subscribeButton: { width: 32, height: 32, borderRadius: 16, alignItems: 'center', justifyContent: 'center' },
  subscribeButtonText: { fontSize: 16, fontWeight: '600' },
  expandHint: { paddingVertical: 10, alignItems: 'center' },
  expandHintText: { fontSize: 12 },
  emptyContainer: { padding: 32, alignItems: 'center', marginTop: 40 },
  emptyIcon: { fontSize: 48, marginBottom: 16 },
  emptyTitle: { fontSize: 20, fontWeight: '600', marginBottom: 8 },
  emptyText: { fontSize: 16, textAlign: 'center', lineHeight: 22 },
});
```

- [ ] **Step 6: Create PromoCardBlock**

New component — tappable card with title, subtitle, and link, driven entirely by config.

```typescript
// src/components/blocks/PromoCardBlock.tsx
import React from 'react';
import { Text, TouchableOpacity, StyleSheet, Linking } from 'react-native';
import { BlockProps } from './BlockRenderer';

export function PromoCardBlock({ config }: BlockProps) {
  const title = config.title as string;
  const subtitle = config.subtitle as string;
  const url = config.url as string;
  const backgroundColor = (config.backgroundColor as string) ?? '#1a1a1a';
  const textColor = (config.textColor as string) ?? '#FFFFFF';

  if (!title || !url) return null;

  return (
    <TouchableOpacity
      style={[styles.container, { backgroundColor }]}
      onPress={() => Linking.openURL(url)}
    >
      <Text style={[styles.title, { color: textColor }]}>{title}</Text>
      {subtitle && (
        <Text style={[styles.subtitle, { color: textColor, opacity: 0.7 }]}>{subtitle}</Text>
      )}
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  container: { padding: 20, marginHorizontal: 16, marginTop: 16, borderRadius: 12 },
  title: { fontSize: 17, fontWeight: '700' },
  subtitle: { fontSize: 14, marginTop: 4 },
});
```

- [ ] **Step 7: Create UserStatsBlock**

Extracted from `ProfileScreen.tsx` — the stats grid showing claimed, active, and subscribed counts.

```typescript
// src/components/blocks/UserStatsBlock.tsx
import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { BlockProps } from './BlockRenderer';
import { useAppData } from '../../context/AppDataContext';
import { useTheme } from '../../hooks/useTheme';
import { api } from '../../api/client';

export function UserStatsBlock({ config }: BlockProps) {
  const { theme } = useTheme();
  const { colors } = theme;
  const { user, undismissedDeals, subscribedCount } = useAppData();
  const [dealsClaimed, setDealsClaimed] = useState(0);

  useEffect(() => {
    if (user?.id) {
      api.getUserStats(user.id)
        .then(stats => setDealsClaimed(stats.dealsClaimed))
        .catch(err => console.error('Failed to fetch user stats:', err));
    }
  }, [user?.id, undismissedDeals]);

  return (
    <View style={[styles.section, { backgroundColor: colors.surface }]}>
      <Text style={[styles.sectionTitle, { color: colors.text }]}>Your Stats</Text>
      <View style={styles.statsGrid}>
        <View style={[styles.statCard, { backgroundColor: colors.surfaceSecondary }]}>
          <Text style={[styles.statNumber, { color: '#E91E63' }]}>{dealsClaimed}</Text>
          <Text style={[styles.statLabel, { color: colors.textMuted }]}>Claimed</Text>
        </View>
        <View style={[styles.statCard, { backgroundColor: colors.surfaceSecondary }]}>
          <Text style={[styles.statNumber, { color: colors.success }]}>{undismissedDeals.length}</Text>
          <Text style={[styles.statLabel, { color: colors.textMuted }]}>Active</Text>
        </View>
        <View style={[styles.statCard, { backgroundColor: colors.surfaceSecondary }]}>
          <Text style={[styles.statNumber, { color: colors.info }]}>{subscribedCount}</Text>
          <Text style={[styles.statLabel, { color: colors.textMuted }]}>Subscribed</Text>
        </View>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  section: { borderRadius: 12, padding: 16, marginBottom: 16 },
  sectionTitle: { fontSize: 18, fontWeight: '600', marginBottom: 16 },
  statsGrid: { flexDirection: 'row', gap: 12 },
  statCard: { flex: 1, padding: 16, borderRadius: 12, alignItems: 'center' },
  statNumber: { fontSize: 32, fontWeight: 'bold' },
  statLabel: { fontSize: 12, marginTop: 4 },
});
```

- [ ] **Step 8: Create SubscriptionListBlock**

Extracted from `ProfileScreen.tsx` — the list of subscribed events with remove buttons.

```typescript
// src/components/blocks/SubscriptionListBlock.tsx
import React, { useMemo } from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { BlockProps } from './BlockRenderer';
import { useAppData } from '../../context/AppDataContext';
import { useTheme } from '../../hooks/useTheme';

export function SubscriptionListBlock({ config }: BlockProps) {
  const { theme } = useTheme();
  const { colors } = theme;
  const { events, isSubscribed, toggleSubscription } = useAppData();

  const subscribedEvents = useMemo(() => {
    return events.filter(e => isSubscribed(e.id));
  }, [events, isSubscribed]);

  if (subscribedEvents.length === 0) return null;

  return (
    <View style={[styles.section, { backgroundColor: colors.surface }]}>
      <Text style={[styles.sectionTitle, { color: colors.text }]}>Your Subscriptions</Text>
      {subscribedEvents.map(event => (
        <View key={event.id} style={[styles.row, { borderBottomColor: colors.border }]}>
          <View style={[styles.teamBadge, { backgroundColor: event.teamColor || colors.accent }]}>
            <Text style={styles.teamBadgeText}>{event.teamId}</Text>
          </View>
          <View style={styles.info}>
            <Text style={[styles.name, { color: colors.text }]} numberOfLines={1}>
              {event.offerName}
            </Text>
            <Text style={[styles.team, { color: colors.textMuted }]}>
              {event.teamName} • {event.partnerName}
            </Text>
          </View>
          <TouchableOpacity
            style={[styles.removeButton, { backgroundColor: colors.surfaceSecondary }]}
            onPress={() => toggleSubscription(event.id)}
          >
            <Text style={[styles.removeText, { color: colors.warning }]}>Remove</Text>
          </TouchableOpacity>
        </View>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  section: { borderRadius: 12, padding: 16, marginBottom: 16 },
  sectionTitle: { fontSize: 18, fontWeight: '600', marginBottom: 16 },
  row: { flexDirection: 'row', alignItems: 'center', paddingVertical: 12, borderBottomWidth: StyleSheet.hairlineWidth },
  teamBadge: { width: 40, height: 40, borderRadius: 20, alignItems: 'center', justifyContent: 'center' },
  teamBadgeText: { color: '#fff', fontWeight: 'bold', fontSize: 12 },
  info: { flex: 1, marginLeft: 12, marginRight: 8 },
  name: { fontSize: 15, fontWeight: '500' },
  team: { fontSize: 12, marginTop: 2 },
  removeButton: { paddingHorizontal: 12, paddingVertical: 6, borderRadius: 12 },
  removeText: { fontSize: 13, fontWeight: '500' },
});
```

- [ ] **Step 9: Create SettingsBlock**

Extracted from `ProfileScreen.tsx` — theme toggle, notification status, account info, and about.

```typescript
// src/components/blocks/SettingsBlock.tsx
import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { BlockProps } from './BlockRenderer';
import { useAppData } from '../../context/AppDataContext';
import { useTheme, ThemeMode } from '../../hooks/useTheme';
import { sendRandomTestNotification } from '../../hooks/usePushNotifications';

export function SettingsBlock({ config }: BlockProps) {
  const { theme, setThemeMode } = useTheme();
  const { colors, mode } = theme;
  const { user, expoPushToken } = useAppData();

  const showThemeToggle = config.showThemeToggle !== false;

  const themeOptions: { label: string; value: ThemeMode; icon: string }[] = [
    { label: 'Light', value: 'light', icon: '☀️' },
    { label: 'Dark', value: 'dark', icon: '🌙' },
    { label: 'System', value: 'system', icon: '⚙️' },
  ];

  return (
    <>
      {/* Appearance */}
      {showThemeToggle && (
        <View style={[styles.section, { backgroundColor: colors.surface }]}>
          <Text style={[styles.sectionTitle, { color: colors.text }]}>Appearance</Text>
          <View style={styles.themeOptions}>
            {themeOptions.map(option => (
              <TouchableOpacity
                key={option.value}
                style={[
                  styles.themeOption,
                  { backgroundColor: colors.surfaceSecondary },
                  mode === option.value && { backgroundColor: colors.accent },
                ]}
                onPress={() => setThemeMode(option.value)}
              >
                <Text style={styles.themeIcon}>{option.icon}</Text>
                <Text style={[styles.themeLabel, {
                  color: mode === option.value ? '#fff' : colors.text,
                }]}>
                  {option.label}
                </Text>
              </TouchableOpacity>
            ))}
          </View>
        </View>
      )}

      {/* Notifications */}
      <View style={[styles.section, { backgroundColor: colors.surface }]}>
        <Text style={[styles.sectionTitle, { color: colors.text }]}>Notifications</Text>
        <View style={styles.row}>
          <View style={styles.rowText}>
            <Text style={[styles.rowTitle, { color: colors.text }]}>Push Notifications</Text>
            <Text style={[styles.rowSubtitle, { color: colors.textMuted }]}>
              {expoPushToken ? 'Enabled' : 'Not available'}
            </Text>
          </View>
          <View style={[styles.statusBadge, {
            backgroundColor: expoPushToken ? colors.successBackground : colors.warningBackground
          }]}>
            <Text style={[styles.statusText, {
              color: expoPushToken ? colors.success : colors.warning
            }]}>
              {expoPushToken ? 'ON' : 'OFF'}
            </Text>
          </View>
        </View>
        {__DEV__ && (
          <TouchableOpacity
            style={[styles.button, { backgroundColor: colors.surfaceSecondary }]}
            onPress={() => sendRandomTestNotification()}
          >
            <Text style={[styles.buttonText, { color: colors.text }]}>
              🔔 Send Test Notification
            </Text>
          </TouchableOpacity>
        )}
      </View>

      {/* Account */}
      <View style={[styles.section, { backgroundColor: colors.surface }]}>
        <Text style={[styles.sectionTitle, { color: colors.text }]}>Account</Text>
        <View style={styles.row}>
          <View style={styles.rowText}>
            <Text style={[styles.rowTitle, { color: colors.text }]}>User ID</Text>
            <Text style={[styles.rowSubtitle, { color: colors.textMuted }]} numberOfLines={1}>
              {user?.id || 'Not logged in'}
            </Text>
          </View>
        </View>
        {expoPushToken && (
          <View style={styles.row}>
            <View style={styles.rowText}>
              <Text style={[styles.rowTitle, { color: colors.text }]}>Push Token</Text>
              <Text style={[styles.rowSubtitle, { color: colors.textMuted }]} numberOfLines={1}>
                {expoPushToken.slice(0, 30)}...
              </Text>
            </View>
          </View>
        )}
      </View>

      {/* About */}
      <View style={[styles.section, { backgroundColor: colors.surface }]}>
        <Text style={[styles.sectionTitle, { color: colors.text }]}>About</Text>
        <View style={styles.row}>
          <Text style={[styles.rowTitle, { color: colors.text }]}>Version</Text>
          <Text style={[styles.rowValue, { color: colors.textMuted }]}>1.0.0</Text>
        </View>
      </View>

      <View style={styles.footer}>
        <Text style={[styles.footerText, { color: colors.textMuted }]}>
          Freebies - Never miss a free offer
        </Text>
      </View>
    </>
  );
}

const styles = StyleSheet.create({
  section: { borderRadius: 12, padding: 16, marginBottom: 16 },
  sectionTitle: { fontSize: 18, fontWeight: '600', marginBottom: 16 },
  themeOptions: { flexDirection: 'row', gap: 8 },
  themeOption: { flex: 1, padding: 12, borderRadius: 12, alignItems: 'center' },
  themeIcon: { fontSize: 24, marginBottom: 4 },
  themeLabel: { fontSize: 12, fontWeight: '500' },
  row: {
    flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between',
    paddingVertical: 12, borderBottomWidth: StyleSheet.hairlineWidth, borderBottomColor: 'rgba(0,0,0,0.1)',
  },
  rowText: { flex: 1, marginRight: 12 },
  rowTitle: { fontSize: 16 },
  rowSubtitle: { fontSize: 12, marginTop: 2 },
  rowValue: { fontSize: 16 },
  statusBadge: { paddingHorizontal: 10, paddingVertical: 4, borderRadius: 12 },
  statusText: { fontSize: 12, fontWeight: '600' },
  button: { padding: 14, borderRadius: 12, alignItems: 'center', marginTop: 12 },
  buttonText: { fontSize: 16, fontWeight: '500' },
  footer: { alignItems: 'center', paddingVertical: 24 },
  footerText: { fontSize: 14 },
});
```

- [ ] **Step 4: Verify all block components compile**

Run: `task mobile:serve`

Expected: App builds without errors (blocks aren't wired to screens yet).

- [ ] **Step 5: Commit**

```bash
git add apps/mobile/src/components/blocks/
git commit -m "feat(mobile): Add block components for server-driven UI"
```

---

## Chunk 5: Mobile — Wire Screens to BlockRenderer

### Task 12: Replace screen layouts with BlockRenderer

**Files:**
- Modify: `apps/mobile/src/screens/DealsScreen.tsx`
- Modify: `apps/mobile/src/screens/DiscoverScreen.tsx`
- Modify: `apps/mobile/src/screens/ProfileScreen.tsx`

- [ ] **Step 1: Update DealsScreen**

Replace the hardcoded content with `BlockRenderer`. Keep any screen-level logic that isn't part of
a block (e.g., modal handlers, notification deep-link setup). The screen becomes a thin wrapper:

```typescript
// The screen keeps its modal, refresh handler, etc.
// but delegates content rendering to BlockRenderer
<BlockRenderer
  screenId="deals"
  refreshing={refreshing}
  onRefresh={handleRefresh}
/>
```

Remove the old inline JSX that rendered the deal list directly. The `ActiveDealsBlock` now handles
that.

- [ ] **Step 2: Update DiscoverScreen**

Same pattern — replace the inline league filter + event list with:

```typescript
<BlockRenderer
  screenId="discover"
  refreshing={refreshing}
  onRefresh={handleRefresh}
/>
```

- [ ] **Step 3: Update ProfileScreen**

Same pattern:

```typescript
<BlockRenderer
  screenId="profile"
  refreshing={refreshing}
  onRefresh={handleRefresh}
/>
```

- [ ] **Step 4: Test the app**

Run: `task mobile:serve`, press `i` for iOS simulator.

Expected: App renders identically to before. All 3 tabs show the same content as the hardcoded
layout. Pull-to-refresh works. Deal modals work. Navigation works.

Check: Start backend with `task api:serve`, verify config endpoint is called on app launch (check
server logs for `GET /api/v1/config`).

- [ ] **Step 5: Test a config change**

With the backend running, modify the seed data to test server-driven behavior:

```bash
# Connect to the database and disable a block
sqlite3 freebie.db "UPDATE screen_blocks SET enabled = 0 WHERE key = 'league-filter-bar'"
```

Background the app and bring it back to foreground. The league filter should disappear from the
Discover tab.

Re-enable it:

```bash
sqlite3 freebie.db "UPDATE screen_blocks SET enabled = 1 WHERE key = 'league-filter-bar'"
```

Background/foreground again — filter reappears.

- [ ] **Step 6: Commit**

```bash
git add apps/mobile/src/screens/
git commit -m "feat(mobile): Wire screens to BlockRenderer for server-driven UI"
```

---

## Chunk 6: Feature Flag Integration & Final Testing

### Task 13: Apply feature flags to event filtering

**Files:**
- Modify: `apps/mobile/src/components/blocks/EventListBlock.tsx` (or wherever events are filtered)

- [ ] **Step 1: Filter events by league flags**

In the `EventListBlock`, read feature flags from `useAppConfig()` and filter the events list before
rendering. Map league names to flag keys:

```typescript
const { config } = useAppConfig();

const leagueFlags: Record<string, string> = {
  MLB: 'enable_mlb',
  NBA: 'enable_nba',
  NFL: 'enable_nfl',
  NHL: 'enable_nhl',
};

const filteredEvents = events.filter((event) => {
  const flagKey = leagueFlags[event.league];
  return !flagKey || config.features[flagKey] !== false;
});
```

- [ ] **Step 2: Filter leagues in LeagueFilterBlock**

Same pattern — only show league tabs for enabled leagues:

```typescript
const enabledLeagues = leagues.filter((league) => {
  const flagKey = `enable_${league.name.toLowerCase()}`;
  return config.features[flagKey] !== false;
});
```

- [ ] **Step 3: Test feature flag behavior**

With backend running:

```bash
sqlite3 freebie.db "UPDATE feature_flags SET enabled = 0 WHERE key = 'enable_nfl'"
```

Background/foreground app. NFL events and the NFL league tab should disappear.

- [ ] **Step 4: Commit**

```bash
git add apps/mobile/src/components/blocks/
git commit -m "feat(mobile): Apply feature flags to league and event filtering"
```

### Task 14: End-to-end test

- [ ] **Step 1: Clean start**

```bash
task clean
task api:serve  # Terminal 1
task mobile:serve  # Terminal 2
```

- [ ] **Step 2: Verify baseline**

App renders all 3 tabs with default config. Config endpoint returns seed data.

- [ ] **Step 3: Add a promo card via migration**

Create `services/api/internal/db/migrations/010_test_promo_card.sql`:

```sql
-- +goose Up
INSERT INTO screen_blocks (id, screen, type, key, position, enabled, config)
VALUES ('blk_discover_shirts', 'discover', 'promo_card', 'custom-shirts', 3, 1,
  '{"title":"Rep Your Team","subtitle":"Custom gear for real fans","url":"https://shop.bitsugar.io","backgroundColor":"#1a1a1a","textColor":"#FFFFFF"}');

-- +goose Down
DELETE FROM screen_blocks WHERE id = 'blk_discover_shirts';
```

Restart server. Background/foreground app. Promo card appears on Discover tab below events.

- [ ] **Step 4: Toggle the promo card off**

```bash
sqlite3 freebie.db "UPDATE screen_blocks SET enabled = 0 WHERE key = 'custom-shirts'"
```

Background/foreground. Card disappears. Toggle back on — card reappears.

- [ ] **Step 5: Reorder blocks**

```bash
sqlite3 freebie.db "UPDATE screen_blocks SET position = 0 WHERE key = 'custom-shirts'"
```

Background/foreground. Promo card now appears above the league filter.

- [ ] **Step 6: Clean up test migration**

Delete the test migration — it was only used to verify the pattern:

```bash
rm services/api/internal/db/migrations/010_test_promo_card.sql
```
