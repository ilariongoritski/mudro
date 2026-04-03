-- Down migration for 009_comment_model.sql
BEGIN;

DROP INDEX IF EXISTS comment_reactions_emoji_idx;
DROP TABLE IF EXISTS comment_reactions;
DROP INDEX IF EXISTS post_comments_parent_comment_id_idx;
ALTER TABLE post_comments DROP COLUMN IF EXISTS parent_comment_id;

COMMIT;
