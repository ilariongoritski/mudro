-- Down migration for 013_unify_media_reactions.sql
BEGIN;

-- 5. Drop indexes
DROP INDEX IF EXISTS stickers_emoji_alias_idx;
DROP INDEX IF EXISTS media_assets_kind_created_idx;

-- 4. Remove LEGACY comments (no-op structurally, but clean up)
COMMENT ON COLUMN messages.media IS NULL;
COMMENT ON COLUMN post_comments.reactions IS NULL;
COMMENT ON COLUMN post_comments.media IS NULL;
COMMENT ON COLUMN posts.media IS NULL;

-- 3. Drop media_asset_id from reaction tables
ALTER TABLE comment_reactions DROP COLUMN IF EXISTS media_asset_id;
ALTER TABLE post_reactions DROP COLUMN IF EXISTS media_asset_id;
ALTER TABLE message_reactions DROP COLUMN IF EXISTS media_asset_id;
ALTER TABLE comment_user_reactions DROP COLUMN IF EXISTS media_asset_id;
ALTER TABLE post_user_reactions DROP COLUMN IF EXISTS media_asset_id;

-- 2. Drop sticker tables
DROP INDEX IF EXISTS stickers_media_asset_id_idx;
DROP INDEX IF EXISTS stickers_pack_id_idx;
DROP TABLE IF EXISTS stickers;
DROP TABLE IF EXISTS sticker_packs;

-- 1. Drop media_assets added columns
ALTER TABLE media_assets DROP COLUMN IF EXISTS blurhash;
ALTER TABLE media_assets DROP COLUMN IF EXISTS file_size_bytes;
ALTER TABLE media_assets DROP COLUMN IF EXISTS duration_ms;

COMMIT;
