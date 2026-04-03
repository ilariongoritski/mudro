-- MUDRO 011_fix_relationships.sql
-- Исправление связей между таблицами для полноценных Go API

-- ============================================================
-- 1. COMMENTS — привязка к автору-юзеру
-- ============================================================

alter table post_comments
  add column if not exists author_id bigint references users(id) on delete set null;

create index if not exists post_comments_author_id_idx
  on post_comments (author_id);

-- Разрешить нативные комменты
alter table post_comments
  drop constraint if exists post_comments_source_check;

alter table post_comments
  add constraint post_comments_source_check
  check (source in ('vk','tg','native'));

-- ============================================================
-- 2. USER REACTIONS — индивидуальные реакции (кто именно)
-- ============================================================

create table if not exists post_user_reactions (
  post_id bigint not null references posts(id) on delete cascade,
  user_id bigint not null references users(id) on delete cascade,
  emoji text not null check (btrim(emoji) <> ''),
  created_at timestamptz not null default now(),
  primary key (post_id, user_id, emoji)
);

create index if not exists post_user_reactions_user_idx
  on post_user_reactions (user_id, created_at desc);

create table if not exists comment_user_reactions (
  comment_id bigint not null references post_comments(id) on delete cascade,
  user_id bigint not null references users(id) on delete cascade,
  emoji text not null check (btrim(emoji) <> ''),
  created_at timestamptz not null default now(),
  primary key (comment_id, user_id, emoji)
);

create index if not exists comment_user_reactions_user_idx
  on comment_user_reactions (user_id, created_at desc);

-- ============================================================
-- 3. NOTIFICATIONS — полиморфная ссылка на источник
-- ============================================================

alter table notifications
  add column if not exists ref_type text,
  add column if not exists ref_id bigint;

-- kind примеры: 'follow', 'like', 'comment', 'message', 'agent_task'
-- ref_type: 'post', 'comment', 'message', 'conversation', 'agent_task'
-- ref_id: id соответствующей сущности

create index if not exists notifications_ref_idx
  on notifications (ref_type, ref_id);

-- ============================================================
-- 4. MESSAGE MEDIA — через media_assets, как у постов
-- ============================================================

create table if not exists message_media_links (
  message_id bigint not null references messages(id) on delete cascade,
  media_asset_id bigint not null references media_assets(id) on delete cascade,
  position integer not null default 1 check (position > 0),
  created_at timestamptz not null default now(),
  primary key (message_id, position)
);

create index if not exists message_media_links_media_idx
  on message_media_links (media_asset_id);

-- ============================================================
-- 5. POST LIKES — user_id вариант (параллельно account-based)
-- ============================================================

create table if not exists post_user_likes (
  post_id bigint not null references posts(id) on delete cascade,
  user_id bigint not null references users(id) on delete cascade,
  created_at timestamptz not null default now(),
  primary key (post_id, user_id)
);

create index if not exists post_user_likes_user_idx
  on post_user_likes (user_id, created_at desc);

-- ============================================================
-- 6. AGENTS — оператор (user) управляющий агентом
-- ============================================================

alter table agents
  add column if not exists operator_id bigint references users(id) on delete set null;

create index if not exists agents_operator_id_idx on agents (operator_id);

-- ============================================================
-- 7. MESSAGE REACTIONS — реакции на сообщения в мессенджере
-- ============================================================

create table if not exists message_reactions (
  message_id bigint not null references messages(id) on delete cascade,
  user_id bigint not null references users(id) on delete cascade,
  emoji text not null check (btrim(emoji) <> ''),
  created_at timestamptz not null default now(),
  primary key (message_id, user_id, emoji)
);
