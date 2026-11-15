-- +goose Up
-- What a push was for: 'trigger' (the deal unlocked) or 'reminder'. Reminder
-- de-duplication reads it so a user gets at most one reminder per deal per day.
-- Rows written before this column are all 'trigger'.
ALTER TABLE notifications ADD COLUMN kind TEXT NOT NULL DEFAULT 'trigger';
CREATE INDEX IF NOT EXISTS idx_notifications_event_kind ON notifications(triggered_event_id, kind);

-- +goose Down
DROP INDEX IF EXISTS idx_notifications_event_kind;
ALTER TABLE notifications DROP COLUMN kind;
