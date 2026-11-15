-- +goose Up
-- Pixel-art sprites for deals. A sprite is data, not an image file: rows of
-- palette keys plus a key -> hex color map. The app draws them natively and
-- the API can render any sprite to PNG (push attachments, sharing, web), so a
-- deal added from the backend gets its picture without an app release.
--
-- pixels:  JSON array of equal-length strings; '.' is transparent.
-- palette: JSON object mapping each non-'.' key to a "#RRGGBB" color.
CREATE TABLE IF NOT EXISTS sprites (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    pixels TEXT NOT NULL,
    palette TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE events ADD COLUMN sprite_id TEXT REFERENCES sprites(id);

-- Fallback shown for any deal without its own sprite.
INSERT INTO sprites (id, name, width, height, pixels, palette) VALUES (
    'gift', 'Gift', 16, 16,
    '["................","...kk.....kk....","..kyyk...kyyk...","...kyyk.kyyk....","....kkykykk.....","kkkkkkkykkkkkkkk","krrrrrrkyrrrrrrk","kRRRRRRkyRRRRRRk","kkkkkkkkykkkkkkk",".krrrrrkyrrrrrk.",".krrrrrkyrrrrrk.",".krrrrrkyrrrrrk.",".kRRRRRkyRRRRRk.",".kRRRRRkyRRRRRk.",".kkkkkkkkkkkkkk.","................"]',
    '{"k":"#000000","y":"#F8D878","r":"#F83800","R":"#A81000"}'
);

-- +goose Down
ALTER TABLE events DROP COLUMN sprite_id;
DROP TABLE IF EXISTS sprites;
