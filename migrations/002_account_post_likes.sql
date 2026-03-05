-- MUDRO 002_account_post_likes.sql
-- Персональные лайки: один аккаунт может поставить только один лайк на пост.

create table if not exists accounts (
  id bigserial primary key,
  external_id text not null,
  platform text not null default 'local',
  display_name text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique (platform, external_id)
);

create index if not exists accounts_platform_external_id_idx
  on accounts (platform, external_id);

create table if not exists post_account_likes (
  post_id bigint not null references posts(id) on delete cascade,
  account_id bigint not null references accounts(id) on delete cascade,
  created_at timestamptz not null default now(),
  primary key (post_id, account_id)
);

create index if not exists post_account_likes_account_id_created_at_idx
  on post_account_likes (account_id, created_at desc);
