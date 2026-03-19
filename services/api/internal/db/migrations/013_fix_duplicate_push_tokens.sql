-- +goose Up
-- Fix duplicate push tokens: keep newest user per token, transfer subscriptions, clean up

-- Step 1: Transfer subscriptions from old duplicate users to the newest user per push_token.
-- For each duplicate push_token group, the newest user (by created_at) is kept.
INSERT OR IGNORE INTO subscriptions (id, user_id, event_id)
SELECT
    'migrated-' || s.id,
    keeper.id,
    s.event_id
FROM subscriptions s
JOIN users u ON s.user_id = u.id
JOIN (
    SELECT push_token, MAX(created_at) as max_created
    FROM users
    WHERE push_token IS NOT NULL AND push_token != ''
    GROUP BY push_token
    HAVING COUNT(*) > 1
) dupes ON u.push_token = dupes.push_token
JOIN users keeper ON keeper.push_token = dupes.push_token AND keeper.created_at = dupes.max_created
WHERE u.id != keeper.id;

-- Step 2: NULL out push_token on older duplicate users
UPDATE users SET push_token = NULL, updated_at = CURRENT_TIMESTAMP
WHERE id IN (
    SELECT u.id
    FROM users u
    JOIN (
        SELECT push_token, MAX(created_at) as max_created
        FROM users
        WHERE push_token IS NOT NULL AND push_token != ''
        GROUP BY push_token
        HAVING COUNT(*) > 1
    ) dupes ON u.push_token = dupes.push_token
    WHERE u.created_at < dupes.max_created
);

-- Step 3: Add partial unique index to prevent future duplicates
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_push_token_unique
ON users(push_token)
WHERE push_token IS NOT NULL AND push_token != '';

-- +goose Down
DROP INDEX IF EXISTS idx_users_push_token_unique;
