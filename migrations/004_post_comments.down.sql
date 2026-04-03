-- Down migration for 004_post_comments.sql
BEGIN;

DROP INDEX IF EXISTS post_comments_post_id_published_at_idx;
DROP INDEX IF EXISTS post_comments_source_source_comment_id_uq;
DROP TABLE IF EXISTS post_comments;

COMMIT;
