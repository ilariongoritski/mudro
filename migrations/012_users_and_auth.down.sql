-- Down migration for 012_users_and_auth.sql
BEGIN;

-- Revert comment source check
ALTER TABLE post_comments DROP CONSTRAINT IF EXISTS post_comments_source_check;
ALTER TABLE post_comments ADD CONSTRAINT post_comments_source_check CHECK (source IN ('vk', 'tg', 'native'));

-- Drop trigger and function for likes count
DROP TRIGGER IF EXISTS trg_post_user_likes_count ON post_user_likes;
DROP FUNCTION IF EXISTS update_post_user_likes_count();

-- Drop post_user_likes table
DROP TABLE IF EXISTS post_user_likes;

-- Drop indexes
DROP INDEX IF EXISTS users_email_idx;
DROP INDEX IF EXISTS users_username_idx;

-- Drop added columns
ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;
ALTER TABLE users DROP COLUMN IF EXISTS display_name;
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;

COMMIT;
