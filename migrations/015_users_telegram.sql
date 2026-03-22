-- 015_users_telegram.sql
-- Adds Telegram identity fields for Mini App auth bootstrap.

alter table users add column if not exists telegram_id bigint;
alter table users add column if not exists telegram_username text;

create unique index if not exists users_telegram_id_uq
  on users (telegram_id)
  where telegram_id is not null;

