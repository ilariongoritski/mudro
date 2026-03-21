-- MUDRO 011_drop_legacy_jsonb.sql
-- Drop legacy jsonb columns after data has been migrated to normalized tables.
--
-- IMPORTANT: Before applying, verify that all legacy data is in normalized tables:
--   SELECT count(*) FROM posts
--     WHERE media IS NOT NULL AND media != 'null'::jsonb AND media != '[]'::jsonb
--     AND NOT EXISTS (SELECT 1 FROM post_media_links WHERE post_id = posts.id);
--   -- Must return 0
--
--   SELECT count(*) FROM post_comments
--     WHERE media IS NOT NULL AND media != 'null'::jsonb AND media != '[]'::jsonb
--     AND NOT EXISTS (SELECT 1 FROM comment_media_links WHERE comment_id = post_comments.id);
--   -- Must return 0

ALTER TABLE posts DROP COLUMN IF EXISTS media;
ALTER TABLE post_comments DROP COLUMN IF EXISTS media;
ALTER TABLE post_comments DROP COLUMN IF EXISTS reactions;
