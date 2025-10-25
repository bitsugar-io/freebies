-- +goose Up
-- Leagues table (sports leagues)
CREATE TABLE IF NOT EXISTS leagues (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    icon TEXT NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Users table (anonymous users identified by device)
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    device_id TEXT UNIQUE NOT NULL,
    push_token TEXT,
    platform TEXT NOT NULL DEFAULT 'unknown',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Events table (freebie opportunities)
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    offer_id TEXT NOT NULL,
    team_id TEXT NOT NULL,
    team_name TEXT NOT NULL,
    league TEXT NOT NULL DEFAULT 'MLB',
    team_color TEXT,
    icon TEXT,
    partner_name TEXT NOT NULL,
    offer_name TEXT NOT NULL,
    offer_description TEXT NOT NULL,
    trigger_condition TEXT NOT NULL,
    trigger_rule TEXT,
    region_code TEXT,
    offer_url TEXT,
    affiliate_url TEXT,
    affiliate_tagline TEXT,
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Subscriptions (user subscribes to event)
CREATE TABLE IF NOT EXISTS subscriptions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, event_id)
);

-- Triggered events (when a rule fires)
CREATE TABLE IF NOT EXISTS triggered_events (
    id TEXT PRIMARY KEY,
    event_id TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    game_id TEXT,
    triggered_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    payload TEXT
);

-- User dismissals (when user acknowledges or dismisses a deal)
CREATE TABLE IF NOT EXISTS dismissals (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    triggered_event_id TEXT NOT NULL REFERENCES triggered_events(id) ON DELETE CASCADE,
    dismissed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    type TEXT NOT NULL DEFAULT 'got_it',
    UNIQUE(user_id, triggered_event_id)
);

-- Notifications sent
CREATE TABLE IF NOT EXISTS notifications (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    triggered_event_id TEXT NOT NULL REFERENCES triggered_events(id) ON DELETE CASCADE,
    sent_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL DEFAULT 'pending'
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_users_device_id ON users(device_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_event_id ON subscriptions(event_id);
CREATE INDEX IF NOT EXISTS idx_triggered_events_event_id ON triggered_events(event_id);
CREATE INDEX IF NOT EXISTS idx_triggered_events_expires_at ON triggered_events(expires_at);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_dismissals_user_id ON dismissals(user_id);
CREATE INDEX IF NOT EXISTS idx_dismissals_triggered_event_id ON dismissals(triggered_event_id);

-- +goose Down
DROP INDEX IF EXISTS idx_dismissals_triggered_event_id;
DROP INDEX IF EXISTS idx_dismissals_user_id;
DROP INDEX IF EXISTS idx_notifications_user_id;
DROP INDEX IF EXISTS idx_triggered_events_expires_at;
DROP INDEX IF EXISTS idx_triggered_events_event_id;
DROP INDEX IF EXISTS idx_subscriptions_event_id;
DROP INDEX IF EXISTS idx_subscriptions_user_id;
DROP INDEX IF EXISTS idx_users_device_id;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS dismissals;
DROP TABLE IF EXISTS triggered_events;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS leagues;
