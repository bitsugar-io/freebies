-- +goose Up
-- LAFC are in the 2026 MLS Cup Playoffs (home Game 2 on Sat Nov 28). Ono's
-- terms (onohawaiianbbq.com/lafc, checked 2026-11-17) cover "a MLS LAFC home
-- game" with no playoff exclusion, so the deal runs for playoff home games.
-- Switch it off again when LAFC's season ends.
UPDATE events SET is_active = 1 WHERE id = 'event-mls-lafc-ono';
-- The 1.x app hides a league whose flag is off; 2.0 ignores the flags.
UPDATE feature_flags SET enabled = 1 WHERE key = 'enable_mls';

-- +goose Down
UPDATE events SET is_active = 0 WHERE id = 'event-mls-lafc-ono';
UPDATE feature_flags SET enabled = 0 WHERE key = 'enable_mls';
