-- +goose Up
-- San Diego Padres - Petco K-9 Promotion
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active
) VALUES (
    'event-mlb-sd-petco', 'sd-petco', 'SD', 'San Diego Padres', 'mlb', '#2F241D', '🐾',
    'Petco', '25% Off at Petco',
    'When Padres pitchers K-9, score 25% off your total purchase the next day at participating Petco locations with your loyalty membership. Just ask to redeem the "Padres K-9 Promotion" at checkout. Petco Love will also donate $500 to a local animal welfare effort every time the Padres K-9.',
    'Padres pitchers record 9+ strikeouts',
    '{"metric": "strikeouts", "scope": "team_pitchers", "operator": ">=", "value": 9, "redemption_window": "same_day"}',
    'us-ca-sd', 'https://www.petco.com/PadresK9',
    '', '', 1
);

-- San Diego Padres - Jack in the Box
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active
) VALUES (
    'event-mlb-sd-jitb', 'sd-jitb', 'SD', 'San Diego Padres', 'mlb', '#2F241D', '🍔',
    'Jack in the Box', 'Free Jumbo Jack',
    'Get a free Jumbo Jack with a large drink purchase the day after the game when the Padres hit a home run at San Diego Jack in the Box locations or on the Jack App using code GOPADRES26.',
    'Padres hit a home run',
    '{"metric": "home_runs", "scope": "team_batters", "operator": ">=", "value": 1, "redemption_window": "same_day"}',
    'us-ca-sd', 'https://www.mlb.com/padres/apps/partners',
    '', '', 1
);

-- San Diego Padres - McDonald's
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active
) VALUES (
    'event-mlb-sd-mcdonalds', 'sd-mcdonalds', 'SD', 'San Diego Padres', 'mlb', '#2F241D', '🍟',
    'McDonald''s', 'Free Medium Fries',
    'Get a free medium fries with $2 purchase the day after the game when the Padres win! Only available on the app at participating McDonald''s locations.',
    'Padres win',
    '{"metric": "win", "scope": "team", "operator": ">=", "value": 1, "redemption_window": "same_day"}',
    'us-ca-sd', 'https://www.mlb.com/padres/apps/partners',
    '', '', 1
);

-- +goose Down
DELETE FROM events WHERE id IN ('event-mlb-sd-petco', 'event-mlb-sd-jitb', 'event-mlb-sd-mcdonalds');
