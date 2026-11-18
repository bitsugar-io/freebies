-- +goose Up
-- Every off-season team shows the month it returns ("BACK FEB", "BACK MAR"),
-- derived from season_start. Angel City alone had a hand-written "FEB 11",
-- so it read differently from the Galaxy's "FEB" on the same screen.
UPDATE teams SET return_label = NULL WHERE id = 'ACFC';

-- +goose Down
UPDATE teams SET return_label = 'FEB 11' WHERE id = 'ACFC';
