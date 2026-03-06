-- +goose Up
-- Initial leagues data
INSERT INTO leagues (id, name, icon, display_order) VALUES
    ('mlb', 'MLB', '⚾', 1);

-- +goose Down
DELETE FROM leagues WHERE id = 'mlb';
