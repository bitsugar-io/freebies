-- +goose Up
-- NFL Teams and Offers (placeholder for future)

-- +goose Down
DELETE FROM events WHERE league = 'nfl';
