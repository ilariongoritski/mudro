-- Down migration for 011_fix_relationships.sql
BEGIN;

-- 7. MESSAGE REACTIONS
DROP TABLE IF EXISTS message_reactions;

-- 6. AGENTS — operator_id
DROP INDEX IF EXISTS agents_operator_id_idx;
ALTER TABLE agents DROP COLUMN IF EXISTS operator_id;

-- 5. POST USER LIKES
DROP INDEX IF EXISTS post_user_likes_user_idx;
DROP TABLE IF EXISTS post_user_likes;

-- 4. MESSAGE MEDIA LINKS
DROP INDEX IF EXISTS message_media_links_media_idx;
DROP TABLE IF EXISTS message_media_links;

-- 3. NOTIFICATIONS — ref columns
DROP INDEX IF EXISTS notifications_ref_idx;
ALTER TABLE notifications DROP COLUMN IF EXISTS ref_id;
ALTER TABLE notifications DROP COLUMN IF EXISTS ref_type;

-- 2. USER REACTIONS
DROP INDEX IF EXISTS comment_user_reactions_user_idx;
DROP TABLE IF EXISTS comment_user_reactions;
DROP INDEX IF EXISTS post_user_reactions_user_idx;
DROP TABLE IF EXISTS post_user_reactions;

-- 1. COMMENTS — author_id and source check revert
ALTER TABLE post_comments DROP CONSTRAINT IF EXISTS post_comments_source_check;
ALTER TABLE post_comments ADD CONSTRAINT post_comments_source_check CHECK (source IN ('vk','tg'));

DROP INDEX IF EXISTS post_comments_author_id_idx;
ALTER TABLE post_comments DROP COLUMN IF EXISTS author_id;

COMMIT;
