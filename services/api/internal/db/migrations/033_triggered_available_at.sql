-- +goose Up
-- When a deal's code or reward becomes usable, if later than the trigger
-- (trigger_rule "unlock_after_days"). NULL = usable as soon as it triggers.
ALTER TABLE triggered_events ADD COLUMN available_at DATETIME;

-- +goose Down
ALTER TABLE triggered_events DROP COLUMN available_at;
