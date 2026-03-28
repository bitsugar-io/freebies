-- +goose Up
-- Re-enable Dodgers/Jack in the Box for 2026 with updated promo code
UPDATE events SET
    is_active = 1,
    offer_description = 'Get a free Jumbo Jack with purchase of a large drink using promo code GODODGERS26 in the Jack App or in-store. Redeemable at participating LA-area locations the day after the game. Cheese is an additional charge.',
    offer_url = 'https://x.com/Dodgers/status/2037387288746049649'
WHERE id = 'event-mlb-lad-jitb';

-- +goose Down
UPDATE events SET is_active = 0 WHERE id = 'event-mlb-lad-jitb';
