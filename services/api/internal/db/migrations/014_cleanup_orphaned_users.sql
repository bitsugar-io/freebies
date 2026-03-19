-- +goose Up
-- Remove orphaned users: no push token, only stale data from duplicate device registrations

DELETE FROM notifications WHERE user_id IN (
    SELECT id FROM users WHERE push_token IS NULL
);

DELETE FROM dismissals WHERE user_id IN (
    SELECT id FROM users WHERE push_token IS NULL
);

DELETE FROM subscriptions WHERE user_id IN (
    SELECT id FROM users WHERE push_token IS NULL
);

DELETE FROM users WHERE push_token IS NULL;

-- +goose Down
-- Cannot restore deleted users
