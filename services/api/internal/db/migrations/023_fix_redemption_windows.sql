-- +goose Up
-- check-triggers cron runs the morning after the game (evaluates yesterday's
-- games), so `now` inside calculateExpiration is already the redemption day.
-- "next_day" therefore expires one day too late, leaving consecutive-day
-- triggers overlapping. Use "same_day" — same fix as 016 for JITB.

UPDATE events SET trigger_rule = '{"metric": "home_double_plays", "scope": "team_fielders", "operator": ">=", "value": 1, "redemption_window": "same_day"}'
WHERE id = 'event-mlb-lad-habit';

UPDATE events SET trigger_rule = '{"metric": "runs", "scope": "team_batters", "operator": ">=", "value": 7, "redemption_window": "same_day"}'
WHERE id = 'event-mlb-col-tacobell';

UPDATE events SET trigger_rule = '{"metric": "home_runs", "scope": "team_batters", "operator": ">=", "value": 1, "redemption_window": "same_day"}'
WHERE id = 'event-mlb-ari-jitb';

UPDATE events SET trigger_rule = '{"metric": "hits", "scope": "team_batters", "operator": ">=", "value": 10, "redemption_window": "same_day"}'
WHERE id = 'event-mlb-ari-bowlero';

-- +goose Down
UPDATE events SET trigger_rule = '{"metric": "home_double_plays", "scope": "team_fielders", "operator": ">=", "value": 1, "redemption_window": "next_day"}'
WHERE id = 'event-mlb-lad-habit';

UPDATE events SET trigger_rule = '{"metric": "runs", "scope": "team_batters", "operator": ">=", "value": 7, "redemption_window": "next_day"}'
WHERE id = 'event-mlb-col-tacobell';

UPDATE events SET trigger_rule = '{"metric": "home_runs", "scope": "team_batters", "operator": ">=", "value": 1, "redemption_window": "next_day"}'
WHERE id = 'event-mlb-ari-jitb';

UPDATE events SET trigger_rule = '{"metric": "hits", "scope": "team_batters", "operator": ">=", "value": 10, "redemption_window": "next_day"}'
WHERE id = 'event-mlb-ari-bowlero';
