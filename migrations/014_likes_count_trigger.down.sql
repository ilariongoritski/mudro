-- Down migration for 014_likes_count_trigger.sql
BEGIN;

DROP TRIGGER IF EXISTS trg_post_user_likes_count ON post_user_likes;
DROP FUNCTION IF EXISTS update_post_likes_count();

COMMIT;
