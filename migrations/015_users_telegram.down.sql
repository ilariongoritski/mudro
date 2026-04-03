-- Down migration for 015_users_telegram.sql
BEGIN;

DROP INDEX IF EXISTS users_telegram_id_uq;
ALTER TABLE users DROP COLUMN IF EXISTS telegram_username;
ALTER TABLE users DROP COLUMN IF EXISTS telegram_id;

COMMIT;
