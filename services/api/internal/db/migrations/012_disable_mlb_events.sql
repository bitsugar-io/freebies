-- +goose Up
-- Disable MLB events until ready for launch
UPDATE events SET is_active = 0 WHERE league = 'mlb';

-- +goose Down
UPDATE events SET is_active = 1 WHERE league = 'mlb';
