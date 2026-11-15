-- +goose Up
-- Two-tone team colors (#RRGGBB, NES-palette approximations) for the Teams screen tiles.
ALTER TABLE teams ADD COLUMN primary_color TEXT;
ALTER TABLE teams ADD COLUMN secondary_color TEXT;
UPDATE teams SET primary_color = '#4428BC', secondary_color = '#F8B800' WHERE id = 'LAL';
UPDATE teams SET primary_color = '#0058F8', secondary_color = '#F83800' WHERE id = 'LAC';
UPDATE teams SET primary_color = '#0058F8', secondary_color = '#FCFCFC' WHERE id = 'LAD';
UPDATE teams SET primary_color = '#F83800', secondary_color = '#0000BC' WHERE id = 'LAA';
UPDATE teams SET primary_color = '#0000BC', secondary_color = '#F8B800' WHERE id = 'LAR';
UPDATE teams SET primary_color = '#3CBCFC', secondary_color = '#F8B800' WHERE id = 'LACH';
UPDATE teams SET primary_color = '#000000', secondary_color = '#BCBCBC' WHERE id = 'LAK';
UPDATE teams SET primary_color = '#E45C10', secondary_color = '#000000' WHERE id = 'ANA';
UPDATE teams SET primary_color = '#000000', secondary_color = '#F8B800' WHERE id = 'LAFC';
UPDATE teams SET primary_color = '#0000BC', secondary_color = '#F8B800' WHERE id = 'LAG';
UPDATE teams SET primary_color = '#000000', secondary_color = '#F8A4C0' WHERE id = 'ACFC';
UPDATE teams SET primary_color = '#4428BC', secondary_color = '#F8B800' WHERE id = 'LAS';
UPDATE teams SET primary_color = '#A81000', secondary_color = '#F8B800' WHERE id = 'USC';
UPDATE teams SET primary_color = '#0078F8', secondary_color = '#F8B800' WHERE id = 'UCLA';

-- +goose Down
ALTER TABLE teams DROP COLUMN secondary_color;
ALTER TABLE teams DROP COLUMN primary_color;
