-- MUDRO 011_simplify_auth.sql

-- Добавляем колонку username
ALTER TABLE users ADD COLUMN IF NOT EXISTS username text;

-- Заполняем username из email для существующих пользователей
UPDATE users SET username = split_part(email, '@', 1) WHERE username IS NULL;

-- Делаем email необязательным
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;

-- Добавляем уникальность для username
ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE (username);

-- Обновляем таблицу токенов (если используется)
ALTER TABLE auth_tokens ADD COLUMN IF NOT EXISTS username text;
ALTER TABLE auth_tokens ALTER COLUMN email DROP NOT NULL;
