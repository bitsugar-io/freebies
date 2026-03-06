-- +goose Up
-- NHL Teams and Offers (placeholder for future)

-- +goose Down
DELETE FROM events WHERE league = 'nhl';
