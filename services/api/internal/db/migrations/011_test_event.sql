-- +goose Up
-- Smoke test event that always triggers (for verifying cronjob pipeline)

INSERT INTO events (
    id, offer_id, team_id, team_name, league, team_color, icon,
    partner_name, offer_name, offer_description,
    trigger_condition, trigger_rule,
    region_code, offer_url, affiliate_url, affiliate_tagline, is_active
) VALUES (
    'event-test-smoke', 'test-smoke', 'TEST', 'Portland Pines', 'test', '#2D5A27', '🍕',
    'Sal''s Pizzeria', 'Free Slice of Pepperoni',
    'Grab a free slice of pepperoni pizza at any Sal''s Pizzeria location. Show this deal at the counter. Valid the day after the game.',
    'Pines pitchers record 5+ strikeouts',
    '{"metric": "score", "scope": "team", "operator": ">=", "value": 1, "redemption_window": "same_day"}',
    'us-or-pdx', '', '', '', 1
);

-- +goose Down
DELETE FROM triggered_events WHERE event_id = 'event-test-smoke';
DELETE FROM events WHERE id = 'event-test-smoke';
