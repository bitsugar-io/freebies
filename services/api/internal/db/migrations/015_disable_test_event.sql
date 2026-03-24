-- +goose Up
UPDATE events SET is_active = 0 WHERE id = 'event-test-smoke';

-- +goose Down
UPDATE events SET is_active = 1 WHERE id = 'event-test-smoke';
