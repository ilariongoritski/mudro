-- Down migration for 001_init.sql
BEGIN;

DROP TABLE IF EXISTS post_reactions;
DROP INDEX IF EXISTS posts_published_at_id_idx;
DROP INDEX IF EXISTS posts_source_source_post_id_uq;
DROP TABLE IF EXISTS posts;

COMMIT;
