-- Down migration for 007_media_assets.sql
BEGIN;

DROP INDEX IF EXISTS comment_media_links_media_idx;
DROP TABLE IF EXISTS comment_media_links;
DROP INDEX IF EXISTS post_media_links_media_idx;
DROP TABLE IF EXISTS post_media_links;
DROP INDEX IF EXISTS media_assets_kind_idx;
DROP INDEX IF EXISTS media_assets_source_idx;
DROP TABLE IF EXISTS media_assets;

COMMIT;
