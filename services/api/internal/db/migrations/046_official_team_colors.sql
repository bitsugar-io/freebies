-- +goose Up
-- Team tiles, the team stripe and YOU WON rays use each team's official colors
-- (owner decision 2026-11-17), replacing the NES-snapped stand-ins that made
-- every gold the same yellow. accent_color is the tile's middle line for teams
-- with a black color, which otherwise drew black on the black name strip.
-- Sources: team brand pages, teamcolorcodes.com, brandcolorcode.com (LAFC);
-- Angel City publishes no hex for sol rosa, so #F4B6C2 is the closest match.
ALTER TABLE teams ADD COLUMN accent_color TEXT;

UPDATE teams SET primary_color = '#552583', secondary_color = '#FDB927' WHERE id = 'LAL';
UPDATE teams SET primary_color = '#1D428A', secondary_color = '#C8102E' WHERE id = 'LAC';
UPDATE teams SET primary_color = '#003594', secondary_color = '#FFA300' WHERE id = 'LAR';
UPDATE teams SET primary_color = '#0080C6', secondary_color = '#FFC20E' WHERE id = 'LACH';
UPDATE teams SET primary_color = '#111111', secondary_color = '#A2AAAD', accent_color = '#FFFFFF' WHERE id = 'LAK';
UPDATE teams SET primary_color = '#FC4C02', secondary_color = '#000000', accent_color = '#B9975B' WHERE id = 'ANA';
UPDATE teams SET primary_color = '#000000', secondary_color = '#C39E6D', accent_color = '#E31837' WHERE id = 'LAFC';
UPDATE teams SET primary_color = '#00245D', secondary_color = '#FFD200' WHERE id = 'LAG';
UPDATE teams SET primary_color = '#000000', secondary_color = '#F4B6C2', accent_color = '#C6A664' WHERE id = 'ACFC';
UPDATE teams SET primary_color = '#990000', secondary_color = '#FFCC00' WHERE id = 'USC';
UPDATE teams SET primary_color = '#2D68C4', secondary_color = '#F2A900' WHERE id = 'UCLA';
UPDATE teams SET primary_color = '#005A9C', secondary_color = '#FFFFFF' WHERE id = 'LAD';
UPDATE teams SET primary_color = '#BA0021', secondary_color = '#003263' WHERE id = 'LAA';
UPDATE teams SET primary_color = '#702F8A', secondary_color = '#FFC72C' WHERE id = 'LAS';

-- +goose Down
UPDATE teams SET primary_color = '#4428BC', secondary_color = '#F8B800' WHERE id IN ('LAL', 'LAS');
UPDATE teams SET primary_color = '#0058F8', secondary_color = '#F83800' WHERE id = 'LAC';
UPDATE teams SET primary_color = '#0000BC', secondary_color = '#F8B800' WHERE id IN ('LAR', 'LAG');
UPDATE teams SET primary_color = '#3CBCFC', secondary_color = '#F8B800' WHERE id = 'LACH';
UPDATE teams SET primary_color = '#000000', secondary_color = '#BCBCBC' WHERE id = 'LAK';
UPDATE teams SET primary_color = '#E45C10', secondary_color = '#000000' WHERE id = 'ANA';
UPDATE teams SET primary_color = '#000000', secondary_color = '#F8B800' WHERE id = 'LAFC';
UPDATE teams SET primary_color = '#000000', secondary_color = '#F8A4C0' WHERE id = 'ACFC';
UPDATE teams SET primary_color = '#A81000', secondary_color = '#F8B800' WHERE id = 'USC';
UPDATE teams SET primary_color = '#0078F8', secondary_color = '#F8B800' WHERE id = 'UCLA';
UPDATE teams SET primary_color = '#0058F8', secondary_color = '#FCFCFC' WHERE id = 'LAD';
UPDATE teams SET primary_color = '#F83800', secondary_color = '#0000BC' WHERE id = 'LAA';
ALTER TABLE teams DROP COLUMN accent_color;
