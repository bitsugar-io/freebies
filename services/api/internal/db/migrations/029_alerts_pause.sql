-- +goose Up
-- NULL = alerts on. A future timestamp pauses pushes until then; 9999-12-31 = until resumed.
ALTER TABLE users ADD COLUMN alerts_paused_until DATETIME;

-- +goose Down
ALTER TABLE users DROP COLUMN alerts_paused_until;
