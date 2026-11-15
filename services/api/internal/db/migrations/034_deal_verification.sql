-- +goose Up
-- Deal verification, 15 Nov 2026 (docs/pivot/deals-verification.md): every deal
-- below was read on the team's or the sponsor's own page that day. Nothing here
-- uses a deal blog or fan site as a source.
--
-- After this migration the only active deals are the ones that are in season,
-- not in-venue-only, and whose metric a source emits (`deals audit` is OK):
--   event-nhl-ana-mcdonalds   Ducks win at home (NHL source)
--   event-nfl-lar-carlsjr     Rams intercept a pass (NFL source, new)
-- Everything else is inactive until its season (MLB, MLS, Galaxy) or until it
-- can be checked (in-venue deals need an in_venue flag and play-by-play).

-- Ducks / McDonald's. Verified for 2026-27, kept active:
-- https://www.nhl.com/ducks/fans/partner-promotions and
-- https://www.nhl.com/ducks/fans/mcdonalds-ducks-win (checked 2026-11-15).
-- "Valid 1x the day after the Ducks win at home, breakfast hours only": the
-- check runs the morning after the game, so same_day + 10:30 is right.

-- Kings / El Pollo Loco stays inactive: https://www.nhl.com/kings/fans/partner-promotions
-- returned 404 on 2026-11-15, the last archived copy is from Aug 2025, and no
-- 2026-27 Kings or El Pollo Loco page mentions it. No 2026-27 evidence.
-- Rams / Carl's Jr. "Famous Star Picks". Verified (page undated, live during the
-- 2026 season): https://www.therams.com/game-day/promotions (checked 2026-11-15).
-- "When your Los Angeles Rams snag an interception you win a free Famous Star."
-- Code RAMS6617, Carl's Jr. app. The code unlocks 2 days after the game and
-- expires in the app 1 week after the unlock, hence unlock_after_days and
-- valid_days. Any game, home or away. Interceptions come from the NFL source
-- (internal/sources/nfl, metric "interceptions").
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active, sprite_id
) VALUES (
    'event-nfl-lar-carlsjr', 'lar-carlsjr', 'LAR', 'Los Angeles Rams', 'nfl', '', '',
    'Carl''s Jr.', 'Free Famous Star',
    'Get a free Famous Star with Cheese after the Rams intercept a pass in any game, home or away. Use code RAMS6617 in the Carl''s Jr. app or at the counter as a registered rewards member. No purchase necessary. The code unlocks 2 days after the game and expires 1 week later. Participating LA-area Carl''s Jr. only.',
    'Rams intercept a pass',
    '{"metric": "interceptions", "scope": "team", "operator": ">=", "value": 1, "redemption_window": "week", "unlock_after_days": 2, "valid_days": 7}',
    'us-ca-la', 'https://www.therams.com/game-day/promotions', '', '', 1, 'burger'
);
-- Ducks / Chick-fil-A "5 goals at home". Verified for 2026-27:
-- https://www.nhl.com/ducks/fans/partner-promotions and https://www.cfasocal.com/
-- (checked 2026-11-15). In-venue only: the reward is claimed in the Chick-fil-A
-- app with location services on, at Honda Center, until 11:59 PM that night.
-- Inactive: in-venue only; needs in_venue flag + play-by-play. The metric
-- home_goals_scored >= 5 already exists in the NHL source.
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active, sprite_id
) VALUES (
    'event-nhl-ana-chickfila', 'ana-chickfila', 'ANA', 'Anaheim Ducks', 'nhl', '', '',
    'Chick-fil-A', 'Free Chicken Sandwich (at the game)',
    'In-venue only. Everyone at Honda Center gets an Original Chick-fil-A Chicken Sandwich in the Chick-fil-A app when the Ducks score 5 or more goals at home. Claim it at the arena that night with location services on.',
    'Ducks score 5+ goals at home',
    '{"metric": "home_goals_scored", "scope": "team", "operator": ">=", "value": 5, "redemption_window": "same_day"}',
    'us-ca-oc', 'https://www.nhl.com/ducks/fans/partner-promotions', '', '', 0, 'chicken-sandwich'
);
-- Clippers / Chick-fil-A: everyone at Intuit Dome gets an Original Chick-fil-A
-- Sandwich in the app if the visiting team misses 2 consecutive free throws in
-- the 4th quarter. Verified on the sponsor's page (undated):
-- https://www.cfasocal.com/ (checked 2026-11-15).
-- Inactive: in-venue only; needs in_venue flag + play-by-play. The NBA source
-- has no q4_opp_consecutive_ft_misses metric yet.
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active, sprite_id
) VALUES (
    'event-nba-lac-chickfila', 'lac-chickfila', 'LAC', 'LA Clippers', 'nba', '', '',
    'Chick-fil-A', 'Free Chicken Sandwich (at the game)',
    'In-venue only. Everyone at Intuit Dome gets an Original Chick-fil-A Chicken Sandwich in the Chick-fil-A app when the visiting team misses 2 free throws in a row in the 4th quarter.',
    'Visitors miss 2 free throws in a row in the 4th quarter',
    '{"metric": "q4_opp_consecutive_ft_misses", "scope": "team", "operator": ">=", "value": 2, "redemption_window": "same_day"}',
    'us-ca-la', 'https://www.cfasocal.com/', '', '', 0, 'chicken-sandwich'
);
-- LAFC / Chick-fil-A: "When LAFC wins a home game, all fans at BMO Stadium are
-- eligible to receive a free Chick-fil-A Chicken Sandwich ... before leaving the
-- stadium." Verified for 2026: https://www.lafc.com/club/promotions/ and
-- https://www.cfasocal.com/ (checked 2026-11-15).
-- Inactive: in-venue only; needs in_venue flag. Recheck for 2027.
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active, sprite_id
) VALUES (
    'event-mls-lafc-chickfila', 'lafc-chickfila', 'LAFC', 'LAFC', 'mls', '', '',
    'Chick-fil-A', 'Free Chicken Sandwich (at the game)',
    'In-venue only. Fans at BMO Stadium get an Original Chick-fil-A Chicken Sandwich when LAFC win a home game. Claim it in the Chick-fil-A app before leaving the stadium, then use it within 3 business days.',
    'LAFC win at home',
    '{"metric": "home_win", "scope": "team", "operator": "==", "value": 1, "redemption_window": "same_day"}',
    'us-ca-la', 'https://www.lafc.com/club/promotions/', '', '', 0, 'chicken-sandwich'
);
-- Galaxy / NORMS. Verified (page undated; the 2026 regular season has ended):
-- https://www.lagalaxy.com/norms (checked 2026-11-15). "NORMS Rewards members
-- win a FREE stack of Buttermilk Hotcakes when the LA Galaxy scores a goal",
-- home games only, valid 3 days after the goal. Inactive until the 2027 home
-- opener; recheck the page first. Sprite is the gift fallback: a pancakes
-- sprite needs the owner's approval (design skill).
INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active, sprite_id
) VALUES (
    'event-mls-lag-norms', 'lag-norms', 'LAG', 'LA Galaxy', 'mls', '', '',
    'NORMS', 'Free stack of Buttermilk Hotcakes',
    'NORMS Rewards members get a free stack of Buttermilk Hotcakes after the Galaxy score a goal at a home game. Show the NORMS Rewards app at the counter within 3 days, dine-in or to-go at a Southern California NORMS. Not valid on NORMS.com, app orders, delivery or catering. One per person.',
    'Galaxy score at home',
    '{"metric": "home_goals_scored", "scope": "team", "operator": ">=", "value": 1, "redemption_window": "week", "valid_days": 3}',
    'us-ca-la', 'https://www.lagalaxy.com/norms', '', '', 0, 'gift'
);
-- LAFC / Ono Hawaiian BBQ. Verified for 2026, undated, so inactive until the
-- 2027 home opener (recheck both pages first):
-- https://onohawaiianbbq.com/lafc/ and https://www.lafc.com/club/promotions/
-- (checked 2026-11-15). The pages disagree: LAFC says "scores in the first
-- half", Ono says "scores first in the first half". We use Ono's stricter
-- reading, which the MLS source emits as home_scored_first_in_first_half.
-- Valid "the next business day after a home game": redemption_window
-- next_business_day. Code LAFCSCORES.
UPDATE events SET
    is_active = 0,
    trigger_condition = 'LAFC score first, in the first half, at home',
    trigger_rule = '{"metric": "home_scored_first_in_first_half", "scope": "team", "operator": "==", "value": 1, "redemption_window": "next_business_day"}',
    offer_description = 'Get a $5.99 Chicken Plate Lunch at Ono Hawaiian BBQ the next business day after LAFC score the first goal of a home game in the first half. Order online as a registered user with code LAFCSCORES. Limit one per order, per user, per day. Not valid on third-party apps. Participating Greater LA locations.',
    offer_url = 'https://onohawaiianbbq.com/lafc/',
    updated_at = CURRENT_TIMESTAMP
WHERE id = 'event-mls-lafc-ono';
-- Angels / McDonald's "Wins Promotion". Verified for 2026 (the 2026 season has
-- ended, so inactive until 2027 Opening Day; recheck first):
-- https://www.mlb.com/angels/tickets/specials/homestand-partner-offers
-- (checked 2026-11-15). "Whenever the Angels win (at home and on the road), fans
-- get the World Famous medium fries for FREE the following day with a $2 minimum
-- purchase from the McDonald's App." Correction: fries, any win, not McGriddles.
UPDATE events SET
    is_active = 0,
    offer_name = 'Free medium fries',
    offer_description = 'Get free medium fries with a $2 minimum purchase in the McDonald''s app the day after the Angels win, at home or on the road. Valid in the Greater LA and Orange County area.',
    trigger_condition = 'Angels win',
    trigger_rule = '{"metric": "win", "scope": "team", "operator": "==", "value": 1, "redemption_window": "same_day"}',
    sprite_id = 'fries',
    offer_url = 'https://www.mlb.com/angels/tickets/specials/homestand-partner-offers',
    updated_at = CURRENT_TIMESTAMP
WHERE id = 'event-mlb-laa-mcdonalds';

-- Angels / Del Taco: the Angels' Del Taco page (5+ runs at home, archived
-- 2026-03-15) now returns 404 and the partner hub no longer lists it. Inactive
-- until a fresh page appears: https://www.mlb.com/angels/tickets/specials/homestand-partner-offers
-- (checked 2026-11-15).
UPDATE events SET is_active = 0, updated_at = CURRENT_TIMESTAMP WHERE id = 'event-mlb-laa-deltaco';

-- Dodgers: the 2026 season is over, and most of these were never confirmed on a
-- primary page (checked 2026-11-15, see docs/pivot/deals-verification.md):
--   Panda Express  https://www.pandaexpress.com/dodgerswin  (HTTP 403 to our fetcher)
--   Jack in the Box reward and code GODODGERS26 verified on the Dodgers' X post
--                   https://x.com/Dodgers/status/2037387288746049649, trigger not
--   McDonald's      no primary source
--   ampm            no 2026 primary source
--   The Habit       https://www.habitburger.com/dodgers/ (2026 code DODGERS26, ended)
-- All inactive until 2027 Opening Day, then rechecked one by one.
UPDATE events SET is_active = 0, updated_at = CURRENT_TIMESTAMP
WHERE id IN ('event-mlb-lad-jitb', 'event-mlb-lad-panda', 'event-mlb-lad-mcdonalds',
             'event-mlb-lad-ampm', 'event-mlb-lad-habit');

-- "Dodger Dog" is a Dodgers trademark: describe the ampm reward as a hot dog.
UPDATE events SET
    offer_name = 'Free hot dog + 16oz Coke',
    offer_description = 'Get a free hot dog and 16oz Coca-Cola when the Dodgers steal a base at a home game. Scan the barcode in the ampm app to redeem. Limit 1 per game. Valid until midnight the day after.',
    updated_at = CURRENT_TIMESTAMP
WHERE id = 'event-mlb-lad-ampm';

-- +goose Down
DELETE FROM events WHERE id IN (
    'event-nfl-lar-carlsjr', 'event-nhl-ana-chickfila', 'event-nba-lac-chickfila',
    'event-mls-lafc-chickfila', 'event-mls-lag-norms'
);

UPDATE events SET
    is_active = 1,
    offer_name = 'Free Dodger Dog + 16oz Coke',
    offer_description = 'Get a free Dodger Dog and 16oz Coca-Cola when the Dodgers steal a base at a home game. Scan the barcode in the ampm app to redeem. Limit 1 per game. Valid until midnight the day after.'
WHERE id = 'event-mlb-lad-ampm';

UPDATE events SET is_active = 1
WHERE id IN ('event-mlb-lad-jitb', 'event-mlb-lad-panda', 'event-mlb-lad-mcdonalds',
             'event-mlb-lad-habit', 'event-mlb-laa-deltaco');

UPDATE events SET
    is_active = 1,
    offer_name = '$2 McGriddles',
    offer_description = 'Get a $2 Sausage, Egg & Cheese McGriddles (breakfast hours only) on the McDonald''s app the day after the Angels win at home. At participating McDonald''s.',
    trigger_condition = 'Angels win at home',
    trigger_rule = '{"metric": "home_win", "scope": "team", "operator": "==", "value": 1, "redemption_window": "same_day"}',
    sprite_id = 'griddle-sandwich',
    offer_url = ''
WHERE id = 'event-mlb-laa-mcdonalds';

UPDATE events SET
    is_active = 1,
    trigger_condition = 'LAFC score in the first half at home',
    trigger_rule = '{"metric": "home_first_half_goals", "scope": "team", "operator": ">=", "value": 1, "redemption_window": "same_day"}',
    offer_description = 'Get a $5.99 Chicken Plate Lunch at Ono Hawaiian BBQ the next business day after LAFC score a first-half goal at home. Use code LAFCSCORES.',
    offer_url = 'https://www.lafc.com/club/promotions'
WHERE id = 'event-mls-lafc-ono';
