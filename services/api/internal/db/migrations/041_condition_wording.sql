-- +goose Up
-- trigger_condition is shown after "If" (TODAY, follow rows) and "Because" (won
-- deals), so it must read as a lowercase clause: "If the visiting team
-- misses...", "Because the visiting team misses...". "Visitors miss" showed as
-- "If Visitors miss".
UPDATE events SET trigger_condition = 'the visiting team misses 2 free throws in a row in the 4th quarter'
WHERE id = 'event-nba-lac-chickfila';
UPDATE events SET trigger_condition = 'the visiting team misses 2 free throws in a row in the 2nd half at Pauley'
WHERE id = 'event-ncaa-ucla-fatsals-hoops';

-- +goose Down
UPDATE events SET trigger_condition = 'Visitors miss 2 free throws in a row in the 4th quarter'
WHERE id = 'event-nba-lac-chickfila';
UPDATE events SET trigger_condition = 'Visitors miss 2 free throws in a row in the 2nd half at Pauley'
WHERE id = 'event-ncaa-ucla-fatsals-hoops';
