-- 023_user_profile.sql
-- User profile fields, rating, completion, activities log.

BEGIN;

ALTER TABLE users 
  ADD COLUMN IF NOT EXISTS display_name text,
  ADD COLUMN IF NOT EXISTS username text,
  ADD COLUMN IF NOT EXISTS age int CHECK (age IS NULL OR age >= 13),
  ADD COLUMN IF NOT EXISTS bio text,
  ADD COLUMN IF NOT EXISTS social_links jsonb DEFAULT '{}',
  ADD COLUMN IF NOT EXISTS avatar_url text,
  ADD COLUMN IF NOT EXISTS profile_completion int DEFAULT 0,
  ADD COLUMN IF NOT EXISTS rating int DEFAULT 0;

-- Make email nullable for profile flexibility
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS users_username_uq ON users(username) WHERE username IS NOT NULL;

CREATE TABLE IF NOT EXISTS user_activities (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type text NOT NULL,
    ref_id bigint,
    metadata jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_activities_user_created 
  ON user_activities(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_user_activities_type 
  ON user_activities(type, created_at DESC);

COMMIT;
