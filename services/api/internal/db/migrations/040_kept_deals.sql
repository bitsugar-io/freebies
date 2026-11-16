-- +goose Up
-- A deal you won stays yours after you unfollow. Unfollowing deletes the
-- subscription row, so before it does, the API records the live deals that
-- were in that user's WON list here (DeleteSubscription, KeepWonDeals).
-- ListActiveTriggeredEventsForUser reads it.
CREATE TABLE IF NOT EXISTS kept_deals (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    triggered_event_id TEXT NOT NULL REFERENCES triggered_events(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, triggered_event_id)
);

-- +goose Down
DROP TABLE IF EXISTS kept_deals;
