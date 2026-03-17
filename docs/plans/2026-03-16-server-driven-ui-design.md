# Server-Driven UI Design

## Problem

The mobile app currently hardcodes which components render on each screen. To add a promo card,
enable a new league, or rearrange a screen layout, we have to ship an app update through the App
Store. We want the backend to control what the app renders — toggle features, reorder content, and
add new blocks — without shipping a new build.

## Design

### Core Concept: Block-Based Screen Composition

The backend defines each screen as an ordered list of **blocks**. Each block has a type, a position,
an enabled flag, and a JSON config blob. The app ships with a registry of block renderers. On
launch, it fetches the full layout config, loops through the blocks for each screen, and renders
whatever it finds. Unknown block types are silently skipped (forward compatibility).

### API

**`GET /api/v1/config`** — public endpoint (no auth required), returns all screen layouts and global
feature flags in a single request.

```json
{
  "features": {
    "enable_nfl": true,
    "enable_nba": true,
    "enable_mlb": true,
    "enable_nhl": false,
    "show_affiliate_links": false,
    "maintenance_mode": false
  },
  "screens": {
    "deals": [
      {
        "type": "banner",
        "key": "nfl-season-banner",
        "config": {
          "text": "NFL Season is here!",
          "backgroundColor": "#013369",
          "textColor": "#FFFFFF",
          "dismissible": true
        }
      },
      {
        "type": "active_deals",
        "key": "active-deals-list",
        "config": {
          "layout": "list",
          "emptyTitle": "No Active Deals",
          "emptySubtitle": "Deals appear here when your teams trigger offers"
        }
      }
    ],
    "discover": [
      {
        "type": "league_filter",
        "key": "league-filter-bar",
        "config": {}
      },
      {
        "type": "event_list",
        "key": "event-list",
        "config": {
          "groupBy": "team"
        }
      },
      {
        "type": "promo_card",
        "key": "custom-shirts",
        "config": {
          "title": "Rep Your Team",
          "subtitle": "Custom gear for real fans",
          "url": "https://shop.bitsugar.io",
          "imageUrl": null,
          "backgroundColor": "#1a1a1a",
          "textColor": "#FFFFFF"
        }
      },
      {
        "type": "promo_card",
        "key": "fanatics-affiliate",
        "config": {
          "title": "Official Gear",
          "subtitle": "Shop team merchandise",
          "url": "https://fanatics.com/?ref=...",
          "imageUrl": null,
          "backgroundColor": "#E31837",
          "textColor": "#FFFFFF"
        }
      }
    ],
    "profile": [
      {
        "type": "user_stats",
        "key": "user-stats",
        "config": {}
      },
      {
        "type": "subscription_list",
        "key": "subscriptions",
        "config": {}
      },
      {
        "type": "settings",
        "key": "settings",
        "config": {
          "showThemeToggle": true
        }
      }
    ]
  }
}
```

**Key properties per block:**

| Field    | Type   | Description                                                      |
| -------- | ------ | ---------------------------------------------------------------- |
| `type`   | string | Block renderer to use. App skips unknown types.                  |
| `key`    | string | Unique identifier. Used for React keys and analytics.            |
| `config` | object | Arbitrary JSON passed to the renderer. Schema varies by type.    |

The `enabled` flag exists in the database but is filtered server-side — the API only returns enabled
blocks. Omitted fields in config (e.g., `imageUrl`) should be treated as absent by the renderer.

**`features` map:** Simple boolean flags for coarse-grained toggles. The app checks these before
rendering league-specific content, affiliate links, etc. These are separate from blocks because they
control cross-cutting behavior (e.g., `enable_nfl` affects event filtering, not just one block).

### Database Schema

New migration adds two tables:

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS feature_flags (
    key TEXT PRIMARY KEY,
    enabled INTEGER NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS screen_blocks (
    id TEXT PRIMARY KEY,
    screen TEXT NOT NULL,          -- 'deals', 'discover', 'profile'
    type TEXT NOT NULL,            -- block renderer type
    key TEXT NOT NULL UNIQUE,      -- unique block identifier
    position INTEGER NOT NULL,     -- display order within screen
    enabled INTEGER NOT NULL DEFAULT 1,
    config TEXT NOT NULL DEFAULT '{}',  -- JSON config blob
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_screen_blocks_screen ON screen_blocks(screen);

-- Seed initial feature flags
INSERT INTO feature_flags (key, enabled) VALUES
    ('enable_mlb', 1),
    ('enable_nba', 1),
    ('enable_nfl', 1),
    ('enable_nhl', 0),
    ('show_affiliate_links', 0),
    ('maintenance_mode', 0);

-- Seed initial screen blocks (matches current hardcoded layout)
-- Deals screen
INSERT INTO screen_blocks (id, screen, type, key, position, config) VALUES
    ('blk_deals_active', 'deals', 'active_deals', 'active-deals-list', 1, '{"layout":"list","emptyTitle":"No Active Deals","emptySubtitle":"Deals appear here when your teams trigger offers"}');

-- Discover screen
INSERT INTO screen_blocks (id, screen, type, key, position, config) VALUES
    ('blk_discover_filter', 'discover', 'league_filter', 'league-filter-bar', 1, '{}'),
    ('blk_discover_events', 'discover', 'event_list', 'event-list', 2, '{"groupBy":"team"}');

-- Profile screen
INSERT INTO screen_blocks (id, screen, type, key, position, config) VALUES
    ('blk_profile_stats', 'profile', 'user_stats', 'user-stats', 1, '{}'),
    ('blk_profile_subs', 'profile', 'subscription_list', 'subscriptions', 2, '{}'),
    ('blk_profile_settings', 'profile', 'settings', 'settings', 3, '{"showThemeToggle":true}');

-- +goose Down
DROP INDEX IF EXISTS idx_screen_blocks_screen;
DROP TABLE IF EXISTS screen_blocks;
DROP TABLE IF EXISTS feature_flags;
```

### sqlc Queries

```sql
-- name: ListFeatureFlags :many
SELECT * FROM feature_flags ORDER BY key;

-- name: ListEnabledScreenBlocks :many
SELECT * FROM screen_blocks
WHERE screen = ? AND enabled = 1
ORDER BY position ASC;

-- name: ListAllEnabledScreenBlocks :many
SELECT * FROM screen_blocks
WHERE enabled = 1
ORDER BY screen, position ASC;
```

### Backend Handler

New handler `GetConfig` assembles the response from both tables:

1. Query all feature flags → build `features` map
2. Query all enabled screen blocks → group by screen, build `screens` map
3. Parse each block's `config` from JSON string to object
4. Return combined response

This is a single public endpoint with no auth, so it can be cached aggressively. The handler sets
`Cache-Control: max-age=60` to keep the app from hammering it.

### Block Type Registry (Mobile)

The app maintains a map of block type → React Native component:

```typescript
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
```

Each block component receives `{ config, screenData }` as props. `config` is the JSON blob from the
backend. `screenData` is the shared context (user, events, subscriptions, etc.) from `AppDataContext`.

The screen renderer is simple:

```typescript
function ScreenRenderer({ screenId }: { screenId: string }) {
  const { config } = useAppConfig();
  const blocks = config?.screens[screenId] ?? [];

  return (
    <ScrollView>
      {blocks.map((block) => {
        const Component = BLOCK_REGISTRY[block.type];
        if (!Component) return null;
        return <Component key={block.key} config={block.config} />;
      })}
    </ScrollView>
  );
}
```

### Initial Block Types

These cover all current UI functionality. Each maps to an existing screen section, refactored into a
standalone block component:

| Type                | Screen    | What it renders                                  |
| ------------------- | --------- | ------------------------------------------------ |
| `banner`            | any       | Dismissible banner with text/color               |
| `active_deals`      | deals     | List of active triggered deals                   |
| `league_filter`     | discover  | Horizontal league filter bar                     |
| `event_list`        | discover  | Grouped event cards (filterable by league)        |
| `promo_card`        | any       | CTA card with title, subtitle, link, colors      |
| `user_stats`        | profile   | Deals claimed / subscriptions count              |
| `subscription_list` | profile   | List of subscribed events                        |
| `settings`          | profile   | Theme toggle, test notification (dev), etc.      |

### Data Flow

```
App Launch + AppState 'active' transition (same pattern as existing data refresh)
  → GET /api/v1/config
  → Store in AppConfigContext
  → Each tab reads its screen blocks from context
  → ScreenRenderer loops blocks, resolves components from registry
  → Each block component pulls data from existing AppDataContext hooks
```

Feature flags are checked at two levels:

1. **Event filtering** — existing `ListEvents` handler already filters by `is_active`. The mobile
   app can additionally filter by league using the `enable_*` flags.
2. **Block visibility** — blocks have `enabled` in the database. Disabled blocks aren't returned by
   the API at all.

### Adding New Content (Workflow)

**Toggle a feature (e.g., enable NHL):**

```sql
-- +goose Up
UPDATE feature_flags SET enabled = 1, updated_at = CURRENT_TIMESTAMP WHERE key = 'enable_nhl';
-- +goose Down
UPDATE feature_flags SET enabled = 0, updated_at = CURRENT_TIMESTAMP WHERE key = 'enable_nhl';
```

**Add a promo card:**

```sql
-- +goose Up
INSERT INTO screen_blocks (id, screen, type, key, position, enabled, config)
VALUES ('blk_discover_fanatics', 'discover', 'promo_card', 'fanatics-affiliate', 3, 1,
  '{"title":"Official Gear","subtitle":"Shop team merchandise","url":"https://fanatics.com/?ref=bitsugar","backgroundColor":"#E31837","textColor":"#FFFFFF"}');
-- +goose Down
DELETE FROM screen_blocks WHERE id = 'blk_discover_fanatics';
```

**Reorder blocks:**

```sql
-- +goose Up
UPDATE screen_blocks SET position = 1 WHERE key = 'event-list';
UPDATE screen_blocks SET position = 2 WHERE key = 'league-filter-bar';
-- +goose Down
UPDATE screen_blocks SET position = 1 WHERE key = 'league-filter-bar';
UPDATE screen_blocks SET position = 2 WHERE key = 'event-list';
```

**Disable a block:**

```sql
-- +goose Up
UPDATE screen_blocks SET enabled = 0 WHERE key = 'custom-shirts';
-- +goose Down
UPDATE screen_blocks SET enabled = 1 WHERE key = 'custom-shirts';
```

### Error Handling

- **Config fetch fails:** App falls back to a hardcoded `DEFAULT_CONFIG` constant in the config
  context, matching the current UI layout. Users never see a broken screen.
- **Unknown block type:** Silently skipped. This means the backend can add new block types before
  the app update ships — they'll just be ignored until the new version is installed.
- **Malformed config:** Block component validates its own config and renders nothing (or a fallback)
  if required fields are missing.

### Testing

- **Backend:** Test `GetConfig` handler returns correct structure. Test that disabled blocks and
  flags are excluded/set correctly.
- **Mobile:** Test `ScreenRenderer` with mock config — correct components rendered, unknown types
  skipped, empty config handled gracefully.
- **Integration:** Verify that toggling a flag or block via migration changes the app behavior on
  next foreground refresh.

### What This Does NOT Do

- **Add new component types without an app update.** New block types require shipping code for the
  renderer. But once a type exists, it can be reused anywhere.
- **Replace a full CMS.** Configs are managed via SQL migrations. An admin UI can be built later on
  the same tables.
- **Change navigation structure.** The 3-tab layout is fixed. Blocks control content within tabs.
