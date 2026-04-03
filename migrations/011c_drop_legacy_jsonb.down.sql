-- Down migration for 011_drop_legacy_jsonb.sql
BEGIN;

-- Re-add the legacy jsonb columns
ALTER TABLE posts ADD COLUMN IF NOT EXISTS media jsonb;
ALTER TABLE post_comments ADD COLUMN IF NOT EXISTS media jsonb;
ALTER TABLE post_comments ADD COLUMN IF NOT EXISTS reactions jsonb;

COMMIT;
