-- Down migration for 010_posts_visibility_and_comments.sql
BEGIN;

-- Drop trigger and function
DROP TRIGGER IF EXISTS trg_post_comments_count ON post_comments;
DROP FUNCTION IF EXISTS update_post_comments_count();

-- Drop visibility index and column
DROP INDEX IF EXISTS posts_visible_published_at_idx;
ALTER TABLE posts DROP COLUMN IF EXISTS visible;

-- Revert comments_count constraints (make nullable again, drop default)
ALTER TABLE posts
  ALTER COLUMN comments_count DROP NOT NULL,
  ALTER COLUMN comments_count DROP DEFAULT;

COMMIT;
