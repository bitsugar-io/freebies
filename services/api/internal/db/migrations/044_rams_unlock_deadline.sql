-- +goose Up
-- Rams / Carl's Jr.: therams.com says "Promotional code available for unlock
-- for 2 days following the qualifying game. Offer expires in app one week from
-- unlock." That is a deadline, not a delay: enter the code within 2 days of the
-- game, then the reward sits in the Carl's Jr. app for a week. It was modeled
-- as "unlocks 2 days after", which told fans to wait when they had to hurry.
UPDATE events SET
    trigger_rule = '{"metric": "interceptions", "scope": "team", "operator": ">=", "value": 1, "redemption_window": "next_day"}',
    offer_description = 'Get a free Famous Star with Cheese with any large fountain drink after the Rams intercept a pass in any game, home or away. Get this week''s code from the Rams promotions page and add it in the Carl''s Jr. app (Rewards > Add Promo Code) as a My Rewards member within 2 days of the game; the burger then stays in your app for a week. Participating LA-area Carl''s Jr. only.'
WHERE id = 'event-nfl-lar-carlsjr';

-- A live win recorded under the old rule: usable now, and the code must be
-- added by the end of day 2 (the old unlock day), not a week later.
UPDATE triggered_events SET
    expires_at = datetime(available_at, '+1 day', '-1 second'),
    available_at = NULL
WHERE event_id = 'event-nfl-lar-carlsjr' AND available_at IS NOT NULL AND expires_at > datetime('now');

-- +goose Down
UPDATE events SET
    trigger_rule = '{"metric": "interceptions", "scope": "team", "operator": ">=", "value": 1, "redemption_window": "week", "unlock_after_days": 2, "valid_days": 7}',
    offer_description = 'Get a free Famous Star with Cheese with any large fountain drink after the Rams intercept a pass in any game, home or away. Get this week''s code from the Rams promotions page and use it in the Carl''s Jr. app or at the counter as a registered rewards member. The code unlocks 2 days after the game and expires 1 week later. Participating LA-area Carl''s Jr. only.'
WHERE id = 'event-nfl-lar-carlsjr';
