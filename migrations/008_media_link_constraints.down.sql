-- Down migration for 008_media_link_constraints.sql
BEGIN;

-- Re-add the unique constraints that were dropped
-- Note: these may fail if duplicates now exist; IF NOT EXISTS not available for constraints
DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_name = 'post_media_links_post_id_media_asset_id_key'
  ) THEN
    ALTER TABLE post_media_links
      ADD CONSTRAINT post_media_links_post_id_media_asset_id_key UNIQUE (post_id, media_asset_id);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_name = 'comment_media_links_comment_id_media_asset_id_key'
  ) THEN
    ALTER TABLE comment_media_links
      ADD CONSTRAINT comment_media_links_comment_id_media_asset_id_key UNIQUE (comment_id, media_asset_id);
  END IF;
END $$;

COMMIT;
