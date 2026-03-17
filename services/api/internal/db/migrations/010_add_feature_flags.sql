-- +goose Up
INSERT INTO feature_flags (key, enabled) VALUES
    ('enable_push_notifications', 1),
    ('enable_subscriptions', 1);

-- +goose Down
DELETE FROM feature_flags WHERE key IN ('enable_push_notifications', 'enable_subscriptions');
