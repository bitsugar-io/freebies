-- +goose Up
-- Add auth token to users table
ALTER TABLE users ADD COLUMN token TEXT;

-- Generate tokens for existing users
UPDATE users SET token = lower(hex(randomblob(32))) WHERE token IS NULL;

-- Make token required and unique for new users
CREATE UNIQUE INDEX idx_users_token ON users(token);

-- +goose Down
DROP INDEX IF EXISTS idx_users_token;
ALTER TABLE users DROP COLUMN token;
