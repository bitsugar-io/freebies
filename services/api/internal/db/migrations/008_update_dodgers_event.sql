-- +goose Up
UPDATE events
SET offer_description = 'For the 2026 Dodgers season, fans can get a free Jumbo Jack at participating L.A.-area Jack in the Box locations the day after the Dodgers pitching staff records seven or more strikeouts. To redeem, use code GODODGERS25 in the Jack in the Box app with a large drink purchase.',
    offer_url = 'https://example.com/dodgers-jitb-details'
WHERE id = 'event-mlb-lad-jitb';

-- +goose Down
UPDATE events
SET offer_description = 'Get a free Jumbo Jack with purchase of a large drink using promo code GODODGERS25 in the Jack App. Redeemable at participating LA-area locations the day after the game.',
    offer_url = 'https://www.jackinthebox.com/offers'
WHERE id = 'event-mlb-lad-jitb';
