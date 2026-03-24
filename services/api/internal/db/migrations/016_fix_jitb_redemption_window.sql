-- +goose Up
UPDATE events SET trigger_rule = '{"metric": "strikeouts", "scope": "team_pitchers", "operator": ">=", "value": 7, "redemption_window": "same_day"}'
WHERE id = 'event-mlb-lad-jitb';

-- +goose Down
UPDATE events SET trigger_rule = '{"metric": "strikeouts", "scope": "team_pitchers", "operator": ">=", "value": 7, "redemption_window": "next_day"}'
WHERE id = 'event-mlb-lad-jitb';
