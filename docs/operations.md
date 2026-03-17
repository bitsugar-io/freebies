# Operations Guide

## Database Access

All commands work with both local SQLite and production Turso.

| Environment | Command prefix |
| ----------- | -------------- |
| Local | `sqlite3 freebie.db` |
| Turso | `turso db shell freebie` |

**Examples use `$DB` as shorthand.** Replace with the command for your environment.

### Reset Database

```bash
# Local — delete and restart server (migrations re-run automatically)
rm freebie.db && task api:serve

# Turso — destroy and recreate
turso db destroy freebie
turso db create freebie
# Then restart the API server to run migrations
```

---

## Feature Flags

Feature flags control app behavior without shipping updates. Changes take effect when users
background/foreground the app or pull-to-refresh.

### View All Flags

```bash
$DB "SELECT key, enabled FROM feature_flags ORDER BY key"
```

### Toggle a Flag

```bash
# Disable
$DB "UPDATE feature_flags SET enabled = 0 WHERE key = 'FLAG_NAME'"

# Enable
$DB "UPDATE feature_flags SET enabled = 1 WHERE key = 'FLAG_NAME'"
```

### Flag Reference

| Flag | Default | Controls |
| ---- | ------- | -------- |
| `enable_mlb` | on | MLB teams, events, and league filter tab |
| `enable_nba` | on | NBA teams, events, and league filter tab |
| `enable_nfl` | on | NFL teams, events, and league filter tab |
| `enable_nhl` | off | NHL teams, events, and league filter tab |
| `show_affiliate_links` | off | Sponsored gear cards (e.g., Fanatics) in deal details |
| `enable_push_notifications` | on | Push notification registration and delivery |
| `enable_subscriptions` | on | Follow/subscribe buttons on events |
| `maintenance_mode` | off | Blocks entire app with "We'll be right back" screen |

### Add a New Flag

Create a migration:

```sql
-- +goose Up
INSERT INTO feature_flags (key, enabled) VALUES ('my_new_flag', 0);
-- +goose Down
DELETE FROM feature_flags WHERE key = 'my_new_flag';
```

Then check it in the app:

```typescript
const { config } = useAppConfig();
if (config.features.my_new_flag === false) return null;
```

---

## Screen Blocks

Screen blocks control what components appear on each screen and in what order. The app fetches
block definitions from `GET /api/v1/config` on launch/foreground.

### View All Blocks

```bash
$DB "SELECT screen, position, type, key, enabled FROM screen_blocks ORDER BY screen, position"
```

### Screens

| Screen ID | Where | Description |
| --------- | ----- | ----------- |
| `deals` | Deals tab | Active triggered deals |
| `discover` | Discover tab | League filter, team list, promo cards |
| `profile` | Profile tab | Stats, subscriptions, settings |
| `deal_detail` | Deal modal | Extra content below deal info (promo cards, etc.) |

### Block Types

| Type | Description | Config options |
| ---- | ----------- | -------------- |
| `banner` | Dismissible banner at top of screen | `text`, `backgroundColor`, `textColor`, `dismissible` |
| `active_deals` | List of active triggered deals | `layout`, `emptyTitle`, `emptySubtitle` |
| `league_filter` | Horizontal league filter pills | (none) |
| `event_list` | Team-grouped event cards | `groupBy` |
| `promo_card` | Tappable CTA card with link | `title`, `subtitle`, `url`, `backgroundColor`, `textColor` |
| `user_stats` | Deals claimed / active / subscribed | (none) |
| `subscription_list` | User's subscribed events with remove | (none) |
| `settings` | Theme, notifications, account info | `showThemeToggle` |

### Add a Block

```bash
$DB "INSERT INTO screen_blocks (id, screen, type, key, position, enabled, config) VALUES (
  'blk_SCREEN_TYPE',     -- unique ID
  'SCREEN_ID',           -- deals, discover, profile, deal_detail
  'BLOCK_TYPE',          -- from block types table above
  'unique-key',          -- unique key for this block
  POSITION,              -- display order (lower = higher on screen)
  1,                     -- 1 = enabled, 0 = disabled
  '{\"key\":\"value\"}'  -- JSON config (see block type options)
)"
```

### Examples

**Add a promo card to the Discover tab:**

```bash
$DB "INSERT INTO screen_blocks (id, screen, type, key, position, enabled, config) VALUES (
  'blk_discover_shirts', 'discover', 'promo_card', 'custom-shirts', 3, 1,
  '{\"title\":\"🛍️ Rep Your Team\",\"subtitle\":\"Custom tees by bitsugar\",\"url\":\"https://shop.bitsugar.io\",\"backgroundColor\":\"#0d3b2e\",\"textColor\":\"#4FF8D2\"}'
)"
```

**Add a promo card to the Deal Detail modal:**

```bash
$DB "INSERT INTO screen_blocks (id, screen, type, key, position, enabled, config) VALUES (
  'blk_deal_shirts', 'deal_detail', 'promo_card', 'deal-custom-shirts', 1, 1,
  '{\"title\":\"🛍️ Rep Your Team\",\"subtitle\":\"Custom tees designed for real fans\",\"url\":\"https://shop.bitsugar.io\",\"backgroundColor\":\"#0d3b2e\",\"textColor\":\"#4FF8D2\"}'
)"
```

**Add a banner to the Deals tab:**

```bash
$DB "INSERT INTO screen_blocks (id, screen, type, key, position, enabled, config) VALUES (
  'blk_deals_banner', 'deals', 'banner', 'season-alert', 0, 1,
  '{\"text\":\"⚾ Dodgers season is live! Follow for free food alerts.\",\"backgroundColor\":\"#005A9C\",\"textColor\":\"#FFFFFF\",\"dismissible\":true}'
)"
```

**Add a promo card to the Profile tab:**

```bash
$DB "INSERT INTO screen_blocks (id, screen, type, key, position, enabled, config) VALUES (
  'blk_profile_merch', 'profile', 'promo_card', 'profile-merch', 0, 1,
  '{\"title\":\"🎁 New Drop\",\"subtitle\":\"Limited edition bitsugar x Dodgers tee\",\"url\":\"https://shop.bitsugar.io\",\"backgroundColor\":\"#2d0a1e\",\"textColor\":\"#FF6B9D\"}'
)"
```

### Disable / Enable a Block

```bash
# Disable
$DB "UPDATE screen_blocks SET enabled = 0 WHERE key = 'custom-shirts'"

# Enable
$DB "UPDATE screen_blocks SET enabled = 1 WHERE key = 'custom-shirts'"
```

### Reorder Blocks

```bash
# Move a block to the top of its screen
$DB "UPDATE screen_blocks SET position = 0 WHERE key = 'custom-shirts'"
```

### Delete a Block

```bash
$DB "DELETE FROM screen_blocks WHERE key = 'custom-shirts'"
```

---

## Common Operations

### Enable a New League

1. Add team/event data via migration (team names, colors, offers)
2. Enable the flag: `$DB "UPDATE feature_flags SET enabled = 1 WHERE key = 'enable_nhl'"`

### Enable Affiliate Links

```bash
$DB "UPDATE feature_flags SET enabled = 1 WHERE key = 'show_affiliate_links'"
```

### Disable Affiliate Links

```bash
$DB "UPDATE feature_flags SET enabled = 0 WHERE key = 'show_affiliate_links'"
```

### Emergency: Take the App Down

```bash
$DB "UPDATE feature_flags SET enabled = 1 WHERE key = 'maintenance_mode'"
```

Users see "We'll be right back." To restore:

```bash
$DB "UPDATE feature_flags SET enabled = 0 WHERE key = 'maintenance_mode'"
```

### Disable Subscriptions During an Outage

```bash
$DB "UPDATE feature_flags SET enabled = 0 WHERE key = 'enable_subscriptions'"
```

### Clean Up Test Blocks

```bash
# Remove all blocks you added for testing (keeps seed data)
$DB "DELETE FROM screen_blocks WHERE id LIKE 'blk_deal_%'"
$DB "DELETE FROM screen_blocks WHERE id LIKE 'blk_discover_shirts%'"
$DB "DELETE FROM screen_blocks WHERE id LIKE 'blk_profile_merch%'"
$DB "DELETE FROM screen_blocks WHERE id LIKE 'blk_deals_banner%'"
```

---

## Config API

`GET /api/v1/config` — public, cached 60s, returns:

```json
{
  "features": { "enable_mlb": true, "maintenance_mode": false, ... },
  "screens": {
    "deals": [{ "type": "active_deals", "key": "...", "config": {} }],
    "discover": [...],
    "deal_detail": [...],
    "profile": [...]
  }
}
```

The app falls back to a hardcoded default config if the endpoint is unreachable.

---

## Local Development

To use local SQLite instead of Turso:

```bash
unset FREEBIE_DATABASE_PATH
task clean
task api:serve
```

The server creates `freebie.db` in the repo root.
