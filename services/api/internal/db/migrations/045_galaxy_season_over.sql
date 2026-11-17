-- +goose Up
-- The Galaxy missed the 2026 MLS Cup Playoffs, so their season is over; the
-- default MLS window (to Dec 10, for playoff teams like LAFC) kept them "in
-- season" with no deals. Back in February: the label comes from season_start.
UPDATE teams SET season_end = '11-10' WHERE id = 'LAG';

-- +goose Down
UPDATE teams SET season_end = '12-10' WHERE id = 'LAG';
