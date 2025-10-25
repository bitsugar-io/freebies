-- +goose Up
-- MLB Teams and Offers

-- Los Angeles Dodgers - Jack in the Box
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active
) VALUES (
    'event-mlb-lad-jitb', 'lad-jitb', 'LAD', 'Los Angeles Dodgers', 'mlb', '#005A9C', '🍔',
    'Jack in the Box', 'Free Jumbo Jack',
    'Get a free Jumbo Jack with purchase of a large drink using promo code GODODGERS25 in the Jack App. Redeemable at participating LA-area locations the day after the game.',
    'Dodgers pitchers record 7+ strikeouts',
    '{"metric": "strikeouts", "scope": "team_pitchers", "operator": ">=", "value": 7, "redemption_window": "next_day"}',
    'us-ca-la', 'https://www.jackinthebox.com/offers',
    'https://www.fanatics.com/mlb/los-angeles-dodgers/o-2329+t-36372287+z-9768-1443776498',
    'Caps, jerseys & championship gear', 1
);

-- +goose Down
DELETE FROM events WHERE league = 'mlb';
