-- 012: extend users table, post_user_likes with trigger, extend comment source

-- Add columns that might not exist yet (idempotent)
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash text;
ALTER TABLE users ADD COLUMN IF NOT EXISTS display_name text;
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url text;

-- Ensure indexes exist
CREATE INDEX IF NOT EXISTS users_username_idx ON users (username);
CREATE INDEX IF NOT EXISTS users_email_idx ON users (email);

-- local user likes (separate from external-platform post_account_likes)
CREATE TABLE IF NOT EXISTS post_user_likes (
  post_id    bigint      NOT NULL REFERENCES posts (id) ON DELETE CASCADE,
  user_id    bigint      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (post_id, user_id)
);

CREATE OR REPLACE FUNCTION update_post_user_likes_count() RETURNS trigger AS $$
BEGIN
  IF tg_op = 'INSERT' THEN
    UPDATE posts SET likes_count = likes_count + 1, updated_at = now()
      WHERE id = new.post_id;
  ELSIF tg_op = 'DELETE' THEN
    UPDATE posts SET likes_count = greatest(likes_count - 1, 0), updated_at = now()
      WHERE id = old.post_id;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_post_user_likes_count ON post_user_likes;
CREATE TRIGGER trg_post_user_likes_count
  AFTER INSERT OR DELETE ON post_user_likes
  FOR EACH ROW EXECUTE FUNCTION update_post_user_likes_count();

-- allow local comments
ALTER TABLE post_comments DROP CONSTRAINT IF EXISTS post_comments_source_check;
ALTER TABLE post_comments ADD CONSTRAINT post_comments_source_check
  CHECK (source IN ('vk', 'tg', 'local'));
