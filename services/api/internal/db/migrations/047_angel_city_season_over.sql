-- +goose Up
-- Angel City lost their NWSL quarterfinal to Kansas City on 2026-11-07, so
-- their season is over (it ran to Nov 25 for a playoff run). Back in
-- February; return_label stays "FEB 11".
UPDATE teams SET season_end = '11-07' WHERE id = 'ACFC';

-- +goose Down
UPDATE teams SET season_end = '11-25' WHERE id = 'ACFC';
