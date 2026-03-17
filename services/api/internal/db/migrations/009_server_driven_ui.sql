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
    ('show_affiliate_links', 1),
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
