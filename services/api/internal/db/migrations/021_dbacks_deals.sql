-- +goose Up
-- Arizona Diamondbacks - Jack in the Box
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active
) VALUES (
    'event-mlb-ari-jitb', 'ari-jitb', 'ARI', 'Arizona Diamondbacks', 'mlb', '#A71930', '🍔',
    'Jack in the Box', 'Free Jumbo Jack',
    'Get a free Jumbo Jack with the purchase of a large drink at your local Jack in the Box the day after any D-backs player hits a home run.',
    'D-backs hit a home run',
    '{"metric": "home_runs", "scope": "team_batters", "operator": ">=", "value": 1, "redemption_window": "next_day"}',
    'us-az', 'https://www.mlb.com/dbacks/fans/in-game-promotions', '', '', 1
);

-- Arizona Diamondbacks - Bowlero
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active
) VALUES (
    'event-mlb-ari-bowlero', 'ari-bowlero', 'ARI', 'Arizona Diamondbacks', 'mlb', '#A71930', '🎳',
    'Bowlero', 'Free Game of Bowling',
    'Get one free game of bowling at Bowlero when the D-backs get 10+ hits. Mention "D-backs 10 hits" at the Bowlero of your choice to redeem.',
    'D-backs get 10+ hits',
    '{"metric": "hits", "scope": "team_batters", "operator": ">=", "value": 10, "redemption_window": "next_day"}',
    'us-az', 'https://www.mlb.com/dbacks/fans/in-game-promotions', '', '', 1
);

-- +goose Down
DELETE FROM events WHERE id IN ('event-mlb-ari-jitb', 'event-mlb-ari-bowlero');
