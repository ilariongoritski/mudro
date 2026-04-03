-- Down migration for 009_users_and_auth.sql
BEGIN;

DROP INDEX IF EXISTS user_sessions_user_id_idx;
DROP TABLE IF EXISTS user_sessions;
DROP INDEX IF EXISTS auth_tokens_email_idx;
DROP TABLE IF EXISTS auth_tokens;
DROP TABLE IF EXISTS users;

COMMIT;
