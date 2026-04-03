-- Down migration for 011_simplify_auth.sql
BEGIN;

-- Revert auth_tokens changes
ALTER TABLE auth_tokens DROP COLUMN IF EXISTS username;
ALTER TABLE auth_tokens ALTER COLUMN email SET NOT NULL;

-- Remove username unique constraint and column
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
ALTER TABLE users DROP COLUMN IF EXISTS username;

-- Make email required again
ALTER TABLE users ALTER COLUMN email SET NOT NULL;

COMMIT;
