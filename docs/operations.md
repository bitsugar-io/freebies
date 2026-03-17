# Operations Guide

## Database Access

Commands in this guide work with both local SQLite and production Turso. Use the matching syntax
for your environment.

**Local (SQLite):** Requires `unset FREEBIE_DATABASE_PATH` so the server uses a local file.

```bash
sqlite3 freebie.db "<SQL>"
```

**Production (Turso):**

```bash
turso db shell freebie "<SQL>"
```

## Feature Flags

Feature flags control app behavior without shipping updates. Changes take effect when users
background and reopen the app (or pull-to-refresh).

### Viewing Current Flags

```bash
# Local
sqlite3 freebie.db "SELECT key, enabled FROM feature_flags ORDER BY key"

# Turso
turso db shell freebie "SELECT key, enabled FROM feature_flags ORDER BY key"
```

### Toggling a Flag

```bash
# Local — disable
sqlite3 freebie.db "UPDATE feature_flags SET enabled = 0 WHERE key = 'enable_nfl'"

# Local — enable
sqlite3 freebie.db "UPDATE feature_flags SET enabled = 1 WHERE key = 'enable_nfl'"

# Turso — disable
turso db shell freebie "UPDATE feature_flags SET enabled = 0 WHERE key = 'enable_nfl'"

# Turso — enable
turso db shell freebie "UPDATE feature_flags SET enabled = 1 WHERE key = 'enable_nfl'"
```

### Available Flags

| Flag | Default | What it controls |
| ---- | ------- | ---------------- |
| `enable_mlb` | on | Show/hide MLB teams, events, and league filter tab |
| `enable_nba` | on | Show/hide NBA teams, events, and league filter tab |
| `enable_nfl` | on | Show/hide NFL teams, events, and league filter tab |
| `enable_nhl` | off | Show/hide NHL teams, events, and league filter tab |
| `show_affiliate_links` | on | Show/hide sponsored gear cards (e.g., Fanatics links) |
| `enable_push_notifications` | on | Enable/disable push notification registration |
| `enable_subscriptions` | on | Enable/disable follow/subscribe buttons |
| `maintenance_mode` | off | Block entire app with "We'll be right back" screen |

### Adding New Flags

Create a goose migration:

```sql
-- +goose Up
INSERT INTO feature_flags (key, enabled) VALUES ('my_new_flag', 0);

-- +goose Down
DELETE FROM feature_flags WHERE key = 'my_new_flag';
```

Then check the flag in the mobile app:

```typescript
const { config } = useAppConfig();
if (config.features.my_new_flag === false) return null;
```

## Screen Blocks

Screen blocks control what components appear on each tab and in what order. The app fetches block
definitions from `GET /api/v1/config` on launch and foreground.

### Viewing Current Blocks

```bash
# Local
sqlite3 freebie.db "SELECT screen, position, type, key, enabled FROM screen_blocks ORDER BY screen, position"

# Turso
turso db shell freebie "SELECT screen, position, type, key, enabled FROM screen_blocks ORDER BY screen, position"
```

### Adding a Block

Example — add a promo card to the Discover tab:

```sql
-- +goose Up
INSERT INTO screen_blocks (id, screen, type, key, position, enabled, config)
VALUES ('blk_discover_shirts', 'discover', 'promo_card', 'custom-shirts', 3, 1,
  '{"title":"Rep Your Team","subtitle":"Custom gear for real fans","url":"https://shop.bitsugar.io","backgroundColor":"#1a1a1a","textColor":"#FFFFFF"}');

-- +goose Down
DELETE FROM screen_blocks WHERE id = 'blk_discover_shirts';
```

### Disabling a Block

```bash
# Local
sqlite3 freebie.db "UPDATE screen_blocks SET enabled = 0 WHERE key = 'custom-shirts'"

# Turso
turso db shell freebie "UPDATE screen_blocks SET enabled = 0 WHERE key = 'custom-shirts'"
```

The API only returns enabled blocks — disabled blocks are invisible to the app.

### Reordering Blocks

```bash
# Local — move promo card above the event list
sqlite3 freebie.db "UPDATE screen_blocks SET position = 0 WHERE key = 'custom-shirts'"

# Turso
turso db shell freebie "UPDATE screen_blocks SET position = 0 WHERE key = 'custom-shirts'"
```

Blocks render in ascending `position` order within each screen.

### Available Block Types

| Type | What it renders | Config options |
| ---- | --------------- | -------------- |
| `banner` | Dismissible banner | `text`, `backgroundColor`, `textColor`, `dismissible` |
| `active_deals` | Active triggered deals list | `layout`, `emptyTitle`, `emptySubtitle` |
| `league_filter` | Horizontal league filter tabs | (none) |
| `event_list` | Team-grouped event cards | `groupBy` |
| `promo_card` | Tappable CTA card with link | `title`, `subtitle`, `url`, `backgroundColor`, `textColor` |
| `user_stats` | Deals claimed / active / subscribed | (none) |
| `subscription_list` | User's subscribed events | (none) |
| `settings` | Theme, notifications, account info | `showThemeToggle` |

### Screens

| Screen ID | Tab |
| --------- | --- |
| `deals` | Active Deals tab |
| `discover` | Discover Events tab |
| `profile` | Profile & Settings tab |

## Common Operations

### Enable a New League

1. Add team/event data via migration (sets team names, colors, offers, etc.)
2. Enable the league flag:

```bash
# Local
sqlite3 freebie.db "UPDATE feature_flags SET enabled = 1 WHERE key = 'enable_nhl'"

# Turso
turso db shell freebie "UPDATE feature_flags SET enabled = 1 WHERE key = 'enable_nhl'"
```

No app update needed — teams appear on next foreground.

### Disable Affiliate / Sponsored Links

```bash
# Local
sqlite3 freebie.db "UPDATE feature_flags SET enabled = 0 WHERE key = 'show_affiliate_links'"

# Turso
turso db shell freebie "UPDATE feature_flags SET enabled = 0 WHERE key = 'show_affiliate_links'"
```

To re-enable:

```bash
# Local
sqlite3 freebie.db "UPDATE feature_flags SET enabled = 1 WHERE key = 'show_affiliate_links'"

# Turso
turso db shell freebie "UPDATE feature_flags SET enabled = 1 WHERE key = 'show_affiliate_links'"
```

### Add a Sponsored Link

1. Add a promo_card block via migration (see "Adding a Block" above)
2. Make sure `show_affiliate_links` is enabled

### Emergency: Take the App Down

```bash
# Local
sqlite3 freebie.db "UPDATE feature_flags SET enabled = 1 WHERE key = 'maintenance_mode'"

# Turso
turso db shell freebie "UPDATE feature_flags SET enabled = 1 WHERE key = 'maintenance_mode'"
```

Users see a "We'll be right back" screen. To restore:

```bash
# Local
sqlite3 freebie.db "UPDATE feature_flags SET enabled = 0 WHERE key = 'maintenance_mode'"

# Turso
turso db shell freebie "UPDATE feature_flags SET enabled = 0 WHERE key = 'maintenance_mode'"
```

### Disable Subscriptions During an Outage

```bash
# Local
sqlite3 freebie.db "UPDATE feature_flags SET enabled = 0 WHERE key = 'enable_subscriptions'"

# Turso
turso db shell freebie "UPDATE feature_flags SET enabled = 0 WHERE key = 'enable_subscriptions'"
```

Follow/unsubscribe buttons become non-functional. Existing subscriptions are preserved.

## Config API

The mobile app calls `GET /api/v1/config` to fetch all flags and blocks. This endpoint is public
(no auth), cached for 60 seconds (`Cache-Control: public, max-age=60`), and returns:

```json
{
  "features": { "enable_mlb": true, "maintenance_mode": false, ... },
  "screens": {
    "deals": [{ "type": "active_deals", "key": "...", "config": {} }],
    "discover": [...],
    "profile": [...]
  }
}
```

The app falls back to a hardcoded default config if the endpoint is unreachable.

## Local Development

To use local SQLite instead of Turso for development:

```bash
unset FREEBIE_DATABASE_PATH
task clean
task api:serve
```

The server creates `freebie.db` in the repo root. You can then use `sqlite3 freebie.db` for all
commands in this guide.
