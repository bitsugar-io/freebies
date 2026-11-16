-- +goose Up
-- LA Freebies: the product is LA-only. Non-LA deals are deactivated (not
-- deleted, so history and triggered_events keep resolving) and LA deals from
-- other sports are seeded. League ids are lowercase to match events.league
-- data (the column default 'MLB' from 001 is never relied on; new rows always
-- set league explicitly).

-- Non-LA events, plus the dewey joke event (a local test deal, not a real offer).
UPDATE events SET is_active = 0
WHERE id IN (
    'event-mlb-sd-petco', 'event-mlb-sd-jitb', 'event-mlb-sd-mcdonalds',
    'event-mlb-col-tacobell',
    'event-mlb-ari-jitb', 'event-mlb-ari-bowlero',
    'event-mlb-lad-dewey'
);

-- Leagues. Only MLB has a game source today; the rest are placeholders so
-- events can be seeded ahead of their sources.
INSERT OR IGNORE INTO leagues (id, name, icon, display_order) VALUES
    ('nba',  'NBA',  '🏀', 2),
    ('nhl',  'NHL',  '🏒', 3),
    ('nfl',  'NFL',  '🏈', 4),
    ('mls',  'MLS',  '⚽', 5),
    ('nwsl', 'NWSL', '⚽', 6),
    ('wnba', 'WNBA', '🏀', 7);

-- enable_nba/nhl/nfl already exist from 009 and are left as they are.
INSERT OR IGNORE INTO feature_flags (key, enabled) VALUES
    ('enable_mls', 0),
    ('enable_nwsl', 0),
    ('enable_wnba', 0);

-- Anaheim Ducks - McDonald's McGriddles. Verified for 2026-27:
-- https://www.nhl.com/ducks/fans/partner-promotions
-- Triggers via the NHL source (internal/sources/nhl, 'home_win' metric).
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active, sprite_id
) VALUES (
    'event-nhl-ana-mcdonalds', 'ana-mcdonalds', 'ANA', 'Anaheim Ducks', 'nhl', '#F47A38', '🍳',
    'McDonald''s', '$2 McGriddles',
    'Get a $2 Sausage, Egg & Cheese McGriddles (breakfast hours only) on the morning after the Ducks win at home. Order in the McDonald''s app at participating McDonald''s.',
    'Ducks win at home',
    '{"metric": "home_win", "scope": "team", "operator": "==", "value": 1, "redemption_window": "same_day", "expires_at_local": "10:30"}',
    'us-ca-oc', 'https://www.nhl.com/ducks/fans/partner-promotions', '', '', 1, 'griddle-sandwich'
);

-- LA Kings - El Pollo Loco. NOT verified for 2026-27, so inactive.
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active, sprite_id
) VALUES (
    'event-nhl-lak-elpolloloco', 'lak-elpolloloco', 'LAK', 'Los Angeles Kings', 'nhl', '#A2AAAD', '🌯',
    'El Pollo Loco', 'Free BRC Burrito + Large Drink',
    'Get a free BRC burrito with the purchase of a large drink after the Kings win at home. Unverified for the 2026-27 season.',
    'Kings win at home',
    '{"metric": "home_win", "scope": "team", "operator": "==", "value": 1, "redemption_window": "same_day"}',
    'us-ca-la', '', '', '', 0, 'burrito'
);

-- LAFC - Ono Hawaiian BBQ. Verified: https://www.lafc.com/club/promotions
-- Code LAFCSCORES. NOTE: needs an MLS Source with a first-half home goal
-- metric; "next business day" is approximated by same_day on the check day.
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active, sprite_id
) VALUES (
    'event-mls-lafc-ono', 'lafc-ono', 'LAFC', 'LAFC', 'mls', '#C39E6D', '🍱',
    'Ono Hawaiian BBQ', '$5.99 Chicken Plate Lunch',
    'Get a $5.99 Chicken Plate Lunch at Ono Hawaiian BBQ the next business day after LAFC score a first-half goal at home. Use code LAFCSCORES.',
    'LAFC score in the first half at home',
    '{"metric": "home_first_half_goals", "scope": "team", "operator": ">=", "value": 1, "redemption_window": "same_day"}',
    'us-ca-la', 'https://www.lafc.com/club/promotions', '', '', 1, 'plate-lunch'
);

-- Any other LA event still without a sprite falls back to the gift.
UPDATE events SET sprite_id = 'gift'
WHERE sprite_id IS NULL AND region_code IN ('us-ca-la', 'us-ca-oc');

-- +goose Down
DELETE FROM events WHERE id IN ('event-nhl-ana-mcdonalds', 'event-nhl-lak-elpolloloco', 'event-mls-lafc-ono');
DELETE FROM feature_flags WHERE key IN ('enable_mls', 'enable_nwsl', 'enable_wnba');
DELETE FROM leagues WHERE id IN ('nba', 'nhl', 'nfl', 'mls', 'nwsl', 'wnba');
UPDATE events SET is_active = 1
WHERE id IN (
    'event-mlb-sd-petco', 'event-mlb-sd-jitb', 'event-mlb-sd-mcdonalds',
    'event-mlb-col-tacobell',
    'event-mlb-ari-jitb', 'event-mlb-ari-bowlero'
);
