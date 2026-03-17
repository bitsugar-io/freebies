-- +goose Up
UPDATE feature_flags SET enabled = 0 WHERE key = 'show_affiliate_links';

-- +goose Down
UPDATE feature_flags SET enabled = 1 WHERE key = 'show_affiliate_links';
