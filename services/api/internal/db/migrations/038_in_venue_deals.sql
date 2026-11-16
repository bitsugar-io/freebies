-- +goose Up
-- At-the-game deals and the second verification pass (15 Nov 2026,
-- docs/pivot/deals-verification.md). Every deal here was read on the team's,
-- the school's or the sponsor's own page; deal blogs were leads only.

-- in_venue: only fans at the game can claim it (the app labels it AT THE GAME).
-- venue: where, for the label ("Intuit Dome"). NULL when not in-venue.
ALTER TABLE events ADD COLUMN in_venue INTEGER NOT NULL DEFAULT 0;
ALTER TABLE events ADD COLUMN venue TEXT;

-- Chick-fil-A at-the-game deals are claimed in the arena the same night
-- (redemption_window "game_night"), so they never send a next-morning alert;
-- the app shows them on game days so fans going know. Ducks and Clippers are
-- in season, so they go active. LAFC stays inactive until 2027.
-- Ducks / Chick-fil-A, 5+ goals at home: https://www.nhl.com/ducks/fans/partner-promotions
UPDATE events SET
    in_venue = 1, venue = 'Honda Center', is_active = 1,
    trigger_rule = '{"metric": "home_goals_scored", "scope": "team", "operator": ">=", "value": 5, "redemption_window": "game_night"}'
WHERE id = 'event-nhl-ana-chickfila';

-- Clippers / Chick-fil-A "Fowl Shot": "If the opposing team misses 2 consecutive
-- free throws in the 4th quarter of Clippers home game, fans at Intuit Dome win
-- a Chick-fil-A Chicken Sandwich in the app ... Claim yours in the app before
-- 11:59p.m. and redeem within 3 business days."
-- https://www.cfasocal.com/promotion-details/fowl-shot (checked 2026-11-15; the
-- creative names Intuit Dome and is (c) 2025, undated otherwise).
UPDATE events SET
    in_venue = 1, venue = 'Intuit Dome', is_active = 1,
    offer_description = 'At the game only. Fans at Intuit Dome win an Original Chick-fil-A Chicken Sandwich when the visiting team misses 2 free throws in a row in the 4th quarter. Location services on in the Chick-fil-A app, claim it before 11:59 PM that night, then use it within 3 business days. One per account per game.',
    trigger_rule = '{"metric": "home_q4_opponent_ft_miss_streak", "scope": "team", "operator": ">=", "value": 2, "redemption_window": "game_night"}',
    offer_url = 'https://www.cfasocal.com/promotion-details/fowl-shot'
WHERE id = 'event-nba-lac-chickfila';

-- LAFC / Chick-fil-A: at the game, claimed before leaving the stadium. Offseason.
UPDATE events SET
    in_venue = 1, venue = 'BMO Stadium',
    trigger_rule = '{"metric": "home_win", "scope": "team", "operator": "==", "value": 1, "redemption_window": "game_night"}'
WHERE id = 'event-mls-lafc-chickfila';

-- Chargers / Firehouse Subs "Cameron Dicker Field Goals For Firehouse Subs":
-- "when Cameron Dicker scores a 40-yard or more field goal during the 2026-2027
-- NFL regular season ... Firehouse Subs Rewards Members get a free medium sub
-- with the purchase of a medium sub, chips, and drink" (the lesser-value sub is
-- free), Firehouse app or FirehouseSubs.com, select LA-area restaurants, valid
-- 3 consecutive days after the game. https://www.firehousesubs.com/offer-terms
-- (2026-27 named; read from the page's indexed text, the page is a JS app).
-- long_field_goal is the team's longest made field goal (any Chargers kicker;
-- Dicker is the only one). Regular season only: deactivate after week 18.
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active, sprite_id
) VALUES (
    'event-nfl-lach-firehouse', 'lach-firehouse', 'LACH', 'Los Angeles Chargers', 'nfl', '', '',
    'Firehouse Subs', 'Free medium sub',
    'Firehouse Subs Rewards members get a free medium sub with the purchase of a medium sub, chips and a drink after Cameron Dicker makes a 40+ yard field goal in a regular-season game. Order in the Firehouse Subs app or at FirehouseSubs.com within 3 days of the game. Select LA-area restaurants.',
    'Dicker makes a 40+ yard field goal',
    '{"metric": "long_field_goal", "scope": "team", "operator": ">=", "value": 40, "redemption_window": "week", "valid_days": 3}',
    'us-ca-la', 'https://www.firehousesubs.com/offer-terms', '', '', 1, 'gift'
);

-- UCLA football / Fat Sal's Deli "Lucky 21": "If the UCLA Bruins score 21 points
-- or more, everyone in attendance wins a FREE order of fries. Show your game
-- ticket at any Fat Sal's Deli location ... valid through following business
-- day." https://uclabruins.com/jmi-promotions (live, undated; Fat Sal's also runs
-- UCLA's 2026 baseball promo). At the game: a ticket is the proof.
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active, sprite_id,
    in_venue, venue
) VALUES (
    'event-ncaa-ucla-fatsals', 'ucla-fatsals', 'UCLA', 'UCLA Bruins', 'ncaa', '', '',
    'Fat Sal''s Deli', 'Free fries (at the game)',
    'At the game only. When UCLA football scores 21 or more points at home, everyone at the Rose Bowl gets a free order of fries. Show your game ticket at any Fat Sal''s Deli through the next business day.',
    'UCLA football scores 21+ at home',
    '{"metric": "football_home_points", "scope": "team", "operator": ">=", "value": 21, "redemption_window": "next_business_day"}',
    'us-ca-la', 'https://uclabruins.com/jmi-promotions', '', '', 1, 'fries',
    1, 'Rose Bowl'
);

-- Lakers / Jack in the Box: "Get 2 FREE tacos on Jack when the Los Angeles
-- Lakers win! Use code LETSGOLAKERS on the Jack App." Jack in the Box on X,
-- 12 Nov 2024: https://x.com/JackBox/status/1856382449686569236. Active: two
-- independent trackers (mealsteals.co, lahomewin.com) logged it firing with the
-- same code after 2026-27 games (22 Oct 2026; a 131-123 win over the Suns, so
-- the old "opponent under 111" cap no longer applies). Recheck in the Jack app.
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active, sprite_id
) VALUES (
    'event-nba-lal-jitb', 'lal-jitb', 'LAL', 'Los Angeles Lakers', 'nba', '', '',
    'Jack in the Box', '2 free tacos',
    'Get 2 free tacos with the purchase of any drink the day after any Lakers win, home or away. Use code LETSGOLAKERS in the Jack in the Box app or online. One code per order, participating restaurants, while supplies last.',
    'Lakers win',
    '{"metric": "win", "scope": "team", "operator": "==", "value": 1, "redemption_window": "same_day"}',
    'us-ca-la', 'https://x.com/JackBox/status/1856382449686569236', '', '', 1, 'taco'
);

-- Kings / McDonald's "McFlurry Minute": a Kings goal in the last minute of the
-- 2nd period gets everyone a free McFlurry with any purchase in the McDonald's
-- app the next day, SoCal. The Kings' redeem page (https://www.nhl.com/kings/fans/redeem,
-- read via its indexed text; it errors when fetched) says "all games this season,
-- home and away"; the Kings posted it on X 27 Mar 2026
-- (https://x.com/LAKings/status/2037581550179828001), and lahomewin.com tracks
-- it for 2026-27. Metric p2_last_minute_goals (NHL source, 19:00+ in the 2nd).
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active, sprite_id
) VALUES (
    'event-nhl-lak-mcdonalds', 'lak-mcdonalds', 'LAK', 'Los Angeles Kings', 'nhl', '', '',
    'McDonald''s', 'Free McFlurry',
    'Get a free McFlurry with any purchase in the McDonald''s app the day after the Kings score in the last minute of the 2nd period, home or away. Find it in Rewards & Deals. Participating Southern California McDonald''s.',
    'Kings score in the last minute of the 2nd period',
    '{"metric": "p2_last_minute_goals", "scope": "team", "operator": ">=", "value": 1, "redemption_window": "same_day"}',
    'us-ca-la', 'https://www.nhl.com/kings/fans/redeem', '', '', 1, 'gift'
);

-- UCLA men's basketball / Fat Sal's "Missed Fry Throws": "If the visiting team
-- misses 2 consecutive free throws in the 2nd half, everyone in attendance will
-- win a free order of regular fries from any Fat Sal's location. Just show your
-- mobile ticket, Fast Break Pass or Den Pass at Fat Sal's on the day following
-- the game." https://uclabruins.com/2025-26-ucla-mens-basketball-themes-promotions
-- (2025-26) and https://uclabruins.com/jmi-promotions (live, undated). No 2026-27
-- page yet; Fat Sal's still runs UCLA promos in 2026 (baseball). At the game.
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active, sprite_id,
    in_venue, venue
) VALUES (
    'event-ncaa-ucla-fatsals-hoops', 'ucla-fatsals-hoops', 'UCLA', 'UCLA Bruins', 'ncaa', '', '',
    'Fat Sal''s Deli', 'Free fries (at the game)',
    'At the game only. When the visiting team misses 2 free throws in a row in the 2nd half of a UCLA men''s basketball home game, everyone at Pauley Pavilion gets a free order of regular fries. Show your mobile ticket, Fast Break Pass or Den Pass at any Fat Sal''s the day after the game.',
    'Visitors miss 2 free throws in a row in the 2nd half at Pauley',
    '{"metric": "basketball_home_h2_opponent_ft_miss_streak", "scope": "team", "operator": ">=", "value": 2, "redemption_window": "same_day"}',
    'us-ca-la', 'https://uclabruins.com/jmi-promotions', '', '', 1, 'fries',
    1, 'Pauley Pavilion'
);

-- +goose Down
DELETE FROM events WHERE id IN ('event-nfl-lach-firehouse', 'event-ncaa-ucla-fatsals', 'event-nba-lal-jitb', 'event-nhl-lak-mcdonalds', 'event-ncaa-ucla-fatsals-hoops');

UPDATE events SET
    is_active = 0,
    trigger_rule = '{"metric": "home_goals_scored", "scope": "team", "operator": ">=", "value": 5, "redemption_window": "same_day"}'
WHERE id = 'event-nhl-ana-chickfila';

UPDATE events SET
    is_active = 0,
    offer_description = 'In-venue only. Everyone at Intuit Dome gets an Original Chick-fil-A Chicken Sandwich in the Chick-fil-A app when the visiting team misses 2 free throws in a row in the 4th quarter.',
    trigger_rule = '{"metric": "q4_opp_consecutive_ft_misses", "scope": "team", "operator": ">=", "value": 2, "redemption_window": "same_day"}',
    offer_url = 'https://www.cfasocal.com/'
WHERE id = 'event-nba-lac-chickfila';

UPDATE events SET
    trigger_rule = '{"metric": "home_win", "scope": "team", "operator": "==", "value": 1, "redemption_window": "same_day"}'
WHERE id = 'event-mls-lafc-chickfila';

ALTER TABLE events DROP COLUMN venue;
ALTER TABLE events DROP COLUMN in_venue;
