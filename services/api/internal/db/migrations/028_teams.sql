-- +goose Up
-- The 14 LA-area teams the Teams screen lists, in season or not.
-- Season windows (MM-DD, may wrap the new year) are approximate; adjust them with a new migration.
-- return_label overrides the "MMM" derived from season_start (e.g. 'FEB 11').
CREATE TABLE teams (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    full_name TEXT NOT NULL,
    league TEXT NOT NULL,
    sport TEXT NOT NULL,
    season_start TEXT NOT NULL,
    season_end TEXT NOT NULL,
    return_label TEXT,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO teams (id, name, full_name, league, sport, season_start, season_end, return_label, display_order) VALUES
    ('LAL', 'Lakers', 'Los Angeles Lakers', 'nba', 'basketball', '10-15', '06-20', NULL, 1),
    ('LAC', 'Clippers', 'LA Clippers', 'nba', 'basketball', '10-15', '06-20', NULL, 2),
    ('LAD', 'Dodgers', 'Los Angeles Dodgers', 'mlb', 'baseball', '03-20', '10-31', NULL, 3),
    ('LAA', 'Angels', 'Los Angeles Angels', 'mlb', 'baseball', '03-20', '10-31', NULL, 4),
    ('LAR', 'Rams', 'Los Angeles Rams', 'nfl', 'football', '09-01', '02-15', NULL, 5),
    ('LACH', 'Chargers', 'Los Angeles Chargers', 'nfl', 'football', '09-01', '02-15', NULL, 6),
    ('LAK', 'Kings', 'Los Angeles Kings', 'nhl', 'hockey', '10-01', '06-20', NULL, 7),
    ('ANA', 'Ducks', 'Anaheim Ducks', 'nhl', 'hockey', '10-01', '06-20', NULL, 8),
    ('LAFC', 'LAFC', 'Los Angeles FC', 'mls', 'soccer', '02-15', '12-10', NULL, 9),
    ('LAG', 'Galaxy', 'LA Galaxy', 'mls', 'soccer', '02-15', '12-10', NULL, 10),
    ('ACFC', 'Angel City', 'Angel City FC', 'nwsl', 'soccer', '02-11', '11-25', 'FEB 11', 11),
    ('LAS', 'Sparks', 'Los Angeles Sparks', 'wnba', 'basketball', '05-15', '10-20', NULL, 12),
    ('USC', 'USC', 'USC Trojans', 'ncaa', 'football', '08-25', '04-10', NULL, 13),
    ('UCLA', 'UCLA', 'UCLA Bruins', 'ncaa', 'football', '08-25', '04-10', NULL, 14);

-- +goose Down
DROP TABLE teams;
