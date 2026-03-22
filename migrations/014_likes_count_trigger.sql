-- MUDRO 014_likes_count_trigger.sql
-- Автоматический подсчёт лайков через триггер

CREATE OR REPLACE FUNCTION update_post_likes_count() RETURNS trigger AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    UPDATE posts SET likes_count = likes_count + 1 WHERE id = NEW.post_id;
    RETURN NEW;
  ELSIF TG_OP = 'DELETE' THEN
    UPDATE posts SET likes_count = likes_count - 1 WHERE id = OLD.post_id;
    RETURN OLD;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_post_user_likes_count ON post_user_likes;

CREATE TRIGGER trg_post_user_likes_count
  AFTER INSERT OR DELETE ON post_user_likes
  FOR EACH ROW EXECUTE FUNCTION update_post_likes_count();
