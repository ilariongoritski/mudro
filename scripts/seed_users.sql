-- ===========================================================
-- Seed: test users for MVP demo
-- Requires: users table with columns (username, email, password_hash, role)
-- Run: psql "postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable" -f scripts/seed_users.sql
-- ===========================================================

-- Bcrypt cost=12:
--   admin123  → $2a$12$KIXFlC5z1eATj7HXJVm5OOqkiSDqI0I.3QsXrZcbGOHx0NHSiWBTe
--   test1234  → $2a$12$8M6vq5D7JhA/gV3cxBfRfeK2A5XkRcLZFQrNgS9bpxHPdvT1I2Oa2

INSERT INTO users (username, email, password_hash, role)
VALUES
  ('admin',    'admin@mudro.local',    '$2a$12$KIXFlC5z1eATj7HXJVm5OOqkiSDqI0I.3QsXrZcbGOHx0NHSiWBTe', 'admin'),
  ('testuser', 'test@mudro.local',     '$2a$12$8M6vq5D7JhA/gV3cxBfRfeK2A5XkRcLZFQrNgS9bpxHPdvT1I2Oa2', 'user')
ON CONFLICT (username) DO UPDATE
  SET email         = EXCLUDED.email,
      password_hash = EXCLUDED.password_hash,
      role          = EXCLUDED.role;

-- Verify
SELECT id, username, email, role FROM users WHERE username IN ('admin', 'testuser');
