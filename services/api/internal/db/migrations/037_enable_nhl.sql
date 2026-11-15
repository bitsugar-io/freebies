-- +goose Up
-- The 1.x app (still in the App Store) hides every league whose enable_<league>
-- flag is false. enable_nhl has been 0 since 009, which hid the Ducks deal from
-- old installs. The 2.x app ignores league flags, so this only affects 1.x.
UPDATE feature_flags SET enabled = 1 WHERE key = 'enable_nhl';

-- +goose Down
UPDATE feature_flags SET enabled = 0 WHERE key = 'enable_nhl';
