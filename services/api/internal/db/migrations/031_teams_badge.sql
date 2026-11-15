-- +goose Up
-- Sport badge sprite per team (sprites table, see 030). Defaults by sport; override per team later.
ALTER TABLE teams ADD COLUMN badge_sprite_id TEXT REFERENCES sprites(id);
UPDATE teams SET badge_sprite_id = 'sport-' || sport;

-- Teams.league_name comes from leagues.name; make sure every team's league has a row.
INSERT OR IGNORE INTO leagues (id, name, icon, display_order) VALUES ('ncaa', 'NCAA', '🎓', 8);

-- +goose Down
DELETE FROM leagues WHERE id = 'ncaa';
ALTER TABLE teams DROP COLUMN badge_sprite_id;
