-- +goose Up
-- Los Angeles Dodgers - The Habit Burger Grill
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active
) VALUES (
    'event-mlb-lad-habit', 'lad-habit', 'LAD', 'Los Angeles Dodgers', 'mlb', '#005A9C', '🍔',
    'The Habit Burger Grill', 'Free Double Char',
    'Get a free Double Char with any $8+ purchase the day after the Dodgers turn a double play at a home game. Enter promo code DODGERS26 in the MyHabit app under "My Offers". Valid at 100+ LA-area locations only. MyHabit membership required (18+, CA resident).',
    'Dodgers turn a double play at home',
    '{"metric": "home_double_plays", "scope": "team_fielders", "operator": ">=", "value": 1, "redemption_window": "next_day"}',
    'us-ca-la', 'https://www.habitburger.com/dodgers/', '', '', 1
);

-- +goose Down
DELETE FROM events WHERE id = 'event-mlb-lad-habit';
