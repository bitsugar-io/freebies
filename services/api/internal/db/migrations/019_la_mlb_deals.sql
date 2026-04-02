-- +goose Up
-- Los Angeles Dodgers - McDonald's McNuggets
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active
) VALUES (
    'event-mlb-lad-mcdonalds', 'lad-mcdonalds', 'LAD', 'Los Angeles Dodgers', 'mlb', '#005A9C', '🍗',
    'McDonald''s', 'Free 6pc McNuggets',
    'Get a free 6pc Chicken McNuggets with a minimum purchase of $2 (excludes tax) on the McDonald''s app the day after the Dodgers score 6+ runs. Must opt in to rewards. Valid 1x/day at participating McDonald''s. Not available during breakfast.',
    'Dodgers score 6+ runs',
    '{"metric": "runs", "scope": "team_batters", "operator": ">=", "value": 6, "redemption_window": "same_day"}',
    'us-ca-la', '', '', '', 1
);

-- Los Angeles Dodgers - ampm Dodger Dog
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active
) VALUES (
    'event-mlb-lad-ampm', 'lad-ampm', 'LAD', 'Los Angeles Dodgers', 'mlb', '#005A9C', '🌭',
    'ampm', 'Free Dodger Dog + 16oz Coke',
    'Get a free Dodger Dog and 16oz Coca-Cola when the Dodgers steal a base at a home game. Scan the barcode in the ampm app to redeem. Limit 1 per game. Valid until midnight the day after.',
    'Dodgers steal a base at home',
    '{"metric": "home_stolen_bases", "scope": "team_batters", "operator": ">=", "value": 1, "redemption_window": "same_day"}',
    'us-ca-la', '', '', '', 1
);

-- Los Angeles Angels - McDonald's McGriddles
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active
) VALUES (
    'event-mlb-laa-mcdonalds', 'laa-mcdonalds', 'LAA', 'Los Angeles Angels', 'mlb', '#BA0021', '🍳',
    'McDonald''s', '$2 McGriddles',
    'Get a $2 Sausage, Egg & Cheese McGriddles (breakfast hours only) on the McDonald''s app the day after the Angels win at home. At participating McDonald''s.',
    'Angels win at home',
    '{"metric": "home_win", "scope": "team", "operator": "==", "value": 1, "redemption_window": "same_day"}',
    'us-ca-la', '', '', '', 1
);

-- Los Angeles Angels - Del Taco
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active
) VALUES (
    'event-mlb-laa-deltaco', 'laa-deltaco', 'LAA', 'Los Angeles Angels', 'mlb', '#BA0021', '🌮',
    'Del Taco', '2 Free Del Tacos',
    'Get two free The Del Tacos with any purchase on the Del Yeah! Rewards app the day after the Angels score 5+ runs at a home game. Valid 10am-11:59pm PT.',
    'Angels score 5+ runs at home',
    '{"metric": "home_runs_scored", "scope": "team_batters", "operator": ">=", "value": 5, "redemption_window": "same_day"}',
    'us-ca-la', '', '', '', 1
);

-- +goose Down
DELETE FROM events WHERE id IN ('event-mlb-lad-mcdonalds', 'event-mlb-lad-ampm', 'event-mlb-laa-mcdonalds', 'event-mlb-laa-deltaco');
