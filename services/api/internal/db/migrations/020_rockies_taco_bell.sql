-- +goose Up
-- Colorado Rockies - Taco Bell
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active
) VALUES (
    'event-mlb-col-tacobell', 'col-tacobell', 'COL', 'Colorado Rockies', 'mlb', '#33006F', '🌮',
    'Taco Bell', '4 Crunchy Tacos for $3',
    'Get 4 Crunchy Tacos for $3 at participating Taco Bell locations the day after the Rockies score 7+ runs. Valid 4-6 PM only.',
    'Rockies score 7+ runs',
    '{"metric": "runs", "scope": "team_batters", "operator": ">=", "value": 7, "redemption_window": "next_day"}',
    'us-co', 'https://www.mlb.com/rockies/ballpark/taco-bell', '', '', 1
);

-- +goose Down
DELETE FROM events WHERE id = 'event-mlb-col-tacobell';
