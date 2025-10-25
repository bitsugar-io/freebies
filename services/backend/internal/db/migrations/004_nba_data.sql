-- +goose Up
-- NBA Teams and Offers (placeholder for future)

-- +goose Down
DELETE FROM events WHERE league = 'nba';
