-- +goose Up
-- LAFC are in the 2026 MLS Cup Playoffs (home Game 2 on Sat Nov 28). Ono's
-- terms (onohawaiianbbq.com/lafc, checked 2026-11-17) cover "a MLS LAFC home
-- game" with no playoff exclusion, so the deal runs for playoff home games.
-- Switch it off again when LAFC's season ends.
UPDATE events SET is_active = 1 WHERE id = 'event-mls-lafc-ono';

-- +goose Down
UPDATE events SET is_active = 0 WHERE id = 'event-mls-lafc-ono';
