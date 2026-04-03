-- Down migration for 002_account_post_likes.sql
BEGIN;

DROP INDEX IF EXISTS post_account_likes_account_id_created_at_idx;
DROP TABLE IF EXISTS post_account_likes;
DROP INDEX IF EXISTS accounts_platform_external_id_idx;
DROP TABLE IF EXISTS accounts;

COMMIT;
