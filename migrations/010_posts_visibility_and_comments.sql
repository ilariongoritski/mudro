-- MUDRO 010_posts_visibility_and_comments.sql
-- 1) Backfill comments_count with actual counts
-- 2) Add visible column (replaces file-based TG visibility filter)
-- 3) Trigger to auto-maintain comments_count

-- Backfill comments_count
UPDATE posts SET comments_count = sub.cnt
FROM (
  SELECT post_id, count(*) AS cnt
  FROM post_comments
  GROUP BY post_id
) sub
WHERE posts.id = sub.post_id
  AND (posts.comments_count IS NULL OR posts.comments_count != sub.cnt);

UPDATE posts SET comments_count = 0
WHERE comments_count IS NULL;

ALTER TABLE posts
  ALTER COLUMN comments_count SET NOT NULL,
  ALTER COLUMN comments_count SET DEFAULT 0;

-- Visibility column
ALTER TABLE posts
  ADD COLUMN IF NOT EXISTS visible boolean NOT NULL DEFAULT true;

CREATE INDEX IF NOT EXISTS posts_visible_published_at_idx
  ON posts (published_at DESC, id DESC)
  WHERE visible = true;

-- Trigger: auto-update comments_count on insert/delete
CREATE OR REPLACE FUNCTION update_post_comments_count() RETURNS trigger AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    UPDATE posts SET comments_count = comments_count + 1, updated_at = now()
    WHERE id = NEW.post_id;
  ELSIF TG_OP = 'DELETE' THEN
    UPDATE posts SET comments_count = GREATEST(comments_count - 1, 0), updated_at = now()
    WHERE id = OLD.post_id;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_post_comments_count ON post_comments;
CREATE TRIGGER trg_post_comments_count
  AFTER INSERT OR DELETE ON post_comments
  FOR EACH ROW EXECUTE FUNCTION update_post_comments_count();
