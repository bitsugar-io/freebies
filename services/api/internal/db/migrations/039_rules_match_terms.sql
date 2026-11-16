-- +goose Up
-- Deal rules re-read against each deal's own terms (15 Nov 2026). The rule
-- language gained three things to say what the terms say (internal/triggers):
-- through_next_business_day, available_from_local and season.

-- UCLA football / Fat Sal's: "valid through following business day". Usable as
-- soon as it triggers; next_business_day wrongly locked it until Monday after a
-- Saturday game.
UPDATE events SET trigger_rule = '{"metric": "football_home_points", "scope": "team", "operator": ">=", "value": 21, "redemption_window": "through_next_business_day"}'
WHERE id = 'event-ncaa-ucla-fatsals';

-- Chargers / Firehouse Subs: "during the 2026-2027 NFL regular season". A
-- playoff field goal must not fire it.
UPDATE events SET trigger_rule = '{"metric": "long_field_goal", "scope": "team", "operator": ">=", "value": 40, "redemption_window": "week", "valid_days": 3, "season": "regular"}'
WHERE id = 'event-nfl-lach-firehouse';

-- Angels / Del Taco: "Valid 10am-11:59pm PT" the day after (offseason, inactive).
UPDATE events SET trigger_rule = '{"metric": "home_runs_scored", "scope": "team_batters", "operator": ">=", "value": 5, "redemption_window": "same_day", "available_from_local": "10:00"}'
WHERE id = 'event-mlb-laa-deltaco';

-- Dodgers / McDonald's: "Not available during breakfast" (breakfast ends 10:30;
-- offseason, inactive).
UPDATE events SET trigger_rule = '{"metric": "runs", "scope": "team_batters", "operator": ">=", "value": 6, "redemption_window": "same_day", "available_from_local": "10:30"}'
WHERE id = 'event-mlb-lad-mcdonalds';

-- Dodgers / The Habit: "during a 2026 regular season Dodgers home game"
-- (offseason, inactive). The MLB source reads gameType from the statsapi
-- schedule, so playoff games are excluded.
UPDATE events SET trigger_rule = '{"metric": "home_double_plays", "scope": "team_fielders", "operator": ">=", "value": 1, "redemption_window": "same_day", "season": "regular"}'
WHERE id = 'event-mlb-lad-habit';

-- +goose Down
UPDATE events SET trigger_rule = '{"metric": "home_double_plays", "scope": "team_fielders", "operator": ">=", "value": 1, "redemption_window": "same_day"}'
WHERE id = 'event-mlb-lad-habit';
UPDATE events SET trigger_rule = '{"metric": "football_home_points", "scope": "team", "operator": ">=", "value": 21, "redemption_window": "next_business_day"}'
WHERE id = 'event-ncaa-ucla-fatsals';
UPDATE events SET trigger_rule = '{"metric": "long_field_goal", "scope": "team", "operator": ">=", "value": 40, "redemption_window": "week", "valid_days": 3}'
WHERE id = 'event-nfl-lach-firehouse';
UPDATE events SET trigger_rule = '{"metric": "home_runs_scored", "scope": "team_batters", "operator": ">=", "value": 5, "redemption_window": "same_day"}'
WHERE id = 'event-mlb-laa-deltaco';
UPDATE events SET trigger_rule = '{"metric": "runs", "scope": "team_batters", "operator": ">=", "value": 6, "redemption_window": "same_day"}'
WHERE id = 'event-mlb-lad-mcdonalds';
