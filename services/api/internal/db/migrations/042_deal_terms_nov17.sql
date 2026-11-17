-- +goose Up
-- Terms checked against the sponsors' pages on 2026-11-17.

-- Rams / Carl's Jr.: therams.com now reads "a free Famous Star with Cheese
-- with any large fountain beverage purchase". The code rotates each game
-- (RAMS459, RAMS7154, RAMS6617...), so point to the page instead of storing one.
UPDATE events SET offer_description = 'Get a free Famous Star with Cheese with any large fountain drink after the Rams intercept a pass in any game, home or away. Get this week''s code from the Rams promotions page and use it in the Carl''s Jr. app or at the counter as a registered rewards member. The code unlocks 2 days after the game and expires 1 week later. Participating LA-area Carl''s Jr. only.'
WHERE id = 'event-nfl-lar-carlsjr';

-- Kings / McDonald's: the offer is "free McFlurry with $2 minimum purchase",
-- and nhl.com/kings/fans/redeem now shows an error page, so drop the link.
UPDATE events SET
    offer_description = 'Get a free McFlurry with a $2 minimum purchase in the McDonald''s app the day after the Kings score in the last minute of the 2nd period, home or away. Find it in Rewards & Deals. Participating Southern California McDonald''s.',
    offer_url = NULL
WHERE id = 'event-nhl-lak-mcdonalds';

-- UCLA hoops / Fat Sal's: uclabruins.com/jmi-promotions says every Fat Sal's
-- offer is "valid through following business day", like the football one.
UPDATE events SET
    offer_description = 'At the game only. When the visiting team misses 2 free throws in a row in the 2nd half of a UCLA men''s basketball home game, everyone at Pauley Pavilion gets a free order of regular fries. Show your mobile ticket, Fast Break Pass or Den Pass at any Fat Sal''s through the next business day.',
    trigger_rule = '{"metric": "basketball_home_h2_opponent_ft_miss_streak", "scope": "team", "operator": ">=", "value": 2, "redemption_window": "through_next_business_day"}'
WHERE id = 'event-ncaa-ucla-fatsals-hoops';

-- +goose Down
UPDATE events SET offer_description = 'Get a free Famous Star with Cheese after the Rams intercept a pass in any game, home or away. Use code RAMS6617 in the Carl''s Jr. app or at the counter as a registered rewards member. No purchase necessary. The code unlocks 2 days after the game and expires 1 week later. Participating LA-area Carl''s Jr. only.'
WHERE id = 'event-nfl-lar-carlsjr';
UPDATE events SET
    offer_description = 'Get a free McFlurry with any purchase in the McDonald''s app the day after the Kings score in the last minute of the 2nd period, home or away. Find it in Rewards & Deals. Participating Southern California McDonald''s.',
    offer_url = 'https://www.nhl.com/kings/fans/redeem'
WHERE id = 'event-nhl-lak-mcdonalds';
UPDATE events SET
    offer_description = 'At the game only. When the visiting team misses 2 free throws in a row in the 2nd half of a UCLA men''s basketball home game, everyone at Pauley Pavilion gets a free order of regular fries. Show your mobile ticket, Fast Break Pass or Den Pass at any Fat Sal''s the day after the game.',
    trigger_rule = '{"metric": "basketball_home_h2_opponent_ft_miss_streak", "scope": "team", "operator": ">=", "value": 2, "redemption_window": "same_day"}'
WHERE id = 'event-ncaa-ucla-fatsals-hoops';
