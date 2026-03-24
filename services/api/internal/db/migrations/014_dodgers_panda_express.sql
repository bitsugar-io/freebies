-- +goose Up
-- Los Angeles Dodgers - Panda Express
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active
) VALUES (
    'event-mlb-lad-panda', 'lad-panda', 'LAD', 'Los Angeles Dodgers', 'mlb', '#005A9C', '🐼',
    'Panda Express', '$7 Panda Plate',
    'Get a $7 Panda Plate using promo code DODGERSWIN in the Panda app or online. Must be a Panda Rewards member. Valid at select participating Panda Express locations. Limit one per account per day.',
    'Dodgers win a home game',
    '{"metric": "home_win", "scope": "team", "operator": "==", "value": 1, "redemption_window": "same_day"}',
    'us-ca-la', 'https://pandaexpress.com/promo/dodgerswin',
    '', '', 1
);

-- +goose Down
DELETE FROM events WHERE id = 'event-mlb-lad-panda';
