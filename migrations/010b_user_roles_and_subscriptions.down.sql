-- Down migration for 010_user_roles_and_subscriptions.sql
BEGIN;

DROP INDEX IF EXISTS user_subscriptions_expires_at_idx;
DROP INDEX IF EXISTS user_subscriptions_user_id_idx;
DROP TABLE IF EXISTS user_subscriptions;

ALTER TABLE users DROP COLUMN IF EXISTS role;

COMMIT;
