-- MUDRO 010_social_messenger_agents.sql
-- Расширение БД: соцсеть + мессенджер + реестр агентов
-- Готовность к микросервисной декомпозиции:
--   feed-svc   -> posts, follows, notifications
--   chat-svc   -> conversations, messages
--   agent-svc  -> agents, agent_queue, agent_task_events
--   auth-svc   -> users, user_sessions, auth_tokens

-- ============================================================
-- 1. PROFILES — расширение users (аватар, био, статус)
-- ============================================================

alter table users
  add column if not exists display_name text,
  add column if not exists avatar_url   text,
  add column if not exists bio          text,
  add column if not exists status       text not null default 'active'
    check (status in ('active','banned','deleted','suspended'));

-- Связь users <-> accounts (один юзер может иметь несколько внешних аккаунтов)
alter table accounts
  add column if not exists user_id bigint references users(id) on delete set null;

create index if not exists accounts_user_id_idx on accounts (user_id);

-- ============================================================
-- 2. FOLLOWS — социальный граф (подписки между юзерами)
-- ============================================================

create table if not exists follows (
  follower_id bigint not null references users(id) on delete cascade,
  following_id bigint not null references users(id) on delete cascade,
  created_at timestamptz not null default now(),
  primary key (follower_id, following_id),
  check (follower_id <> following_id)
);

create index if not exists follows_following_id_idx on follows (following_id);

-- ============================================================
-- 3. NATIVE POSTS — расширение posts для нативного контента
-- ============================================================

-- Разрешить нативные посты (source='native') и привязать автора
alter table posts
  drop constraint if exists posts_source_check;

alter table posts
  add constraint posts_source_check
  check (source in ('vk','tg','native'));

alter table posts
  add column if not exists author_id bigint references users(id) on delete set null;

create index if not exists posts_author_id_idx on posts (author_id);

-- ============================================================
-- 4. CONVERSATIONS & MESSAGES — мессенджер
-- ============================================================

create table if not exists conversations (
  id bigserial primary key,
  kind text not null default 'direct' check (kind in ('direct','group','channel')),
  title text,
  avatar_url text,
  created_by bigint references users(id) on delete set null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists conversation_members (
  conversation_id bigint not null references conversations(id) on delete cascade,
  user_id bigint not null references users(id) on delete cascade,
  role text not null default 'member' check (role in ('owner','admin','member')),
  joined_at timestamptz not null default now(),
  muted_until timestamptz,
  primary key (conversation_id, user_id)
);

create index if not exists conversation_members_user_id_idx
  on conversation_members (user_id);

create table if not exists messages (
  id bigserial primary key,
  conversation_id bigint not null references conversations(id) on delete cascade,
  sender_id bigint references users(id) on delete set null,
  reply_to_id bigint references messages(id) on delete set null,
  body text,
  media jsonb,
  edited_at timestamptz,
  deleted_at timestamptz,
  created_at timestamptz not null default now()
);

create index if not exists messages_conversation_created_idx
  on messages (conversation_id, created_at desc, id desc);

create index if not exists messages_sender_id_idx
  on messages (sender_id);

-- Курсор прочтения: последнее прочитанное сообщение в диалоге
create table if not exists message_reads (
  conversation_id bigint not null references conversations(id) on delete cascade,
  user_id bigint not null references users(id) on delete cascade,
  last_read_message_id bigint not null references messages(id) on delete cascade,
  read_at timestamptz not null default now(),
  primary key (conversation_id, user_id)
);

-- ============================================================
-- 5. NOTIFICATIONS — уведомления
-- ============================================================

create table if not exists notifications (
  id bigserial primary key,
  user_id bigint not null references users(id) on delete cascade,
  kind text not null,
  title text,
  body text,
  payload jsonb not null default '{}'::jsonb,
  is_read boolean not null default false,
  created_at timestamptz not null default now()
);

create index if not exists notifications_user_unread_idx
  on notifications (user_id, created_at desc)
  where is_read = false;

create index if not exists notifications_user_id_created_idx
  on notifications (user_id, created_at desc);

-- ============================================================
-- 6. AGENTS — реестр агентов для управления командой
-- ============================================================

create table if not exists agents (
  id bigserial primary key,
  slug text not null unique,
  display_name text not null,
  kind text not null default 'worker' check (kind in ('worker','reviewer','planner','operator')),
  status text not null default 'idle' check (status in ('idle','busy','offline','error')),
  config jsonb not null default '{}'::jsonb,
  capabilities text[] not null default '{}',
  last_heartbeat_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

-- Привязать задачи в очереди к конкретному агенту
alter table agent_queue
  add column if not exists agent_id bigint references agents(id) on delete set null;

create index if not exists agent_queue_agent_id_idx on agent_queue (agent_id);

-- Привязать события к агенту
alter table agent_task_events
  add column if not exists agent_id bigint references agents(id) on delete set null;

create index if not exists agent_task_events_agent_id_idx on agent_task_events (agent_id);

-- Лог общения агентов (для отладки и аудита команды)
create table if not exists agent_messages (
  id bigserial primary key,
  from_agent_id bigint references agents(id) on delete set null,
  to_agent_id bigint references agents(id) on delete set null,
  task_id bigint references agent_queue(id) on delete set null,
  kind text not null default 'chat' check (kind in ('chat','command','report','error')),
  body text not null,
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);

create index if not exists agent_messages_task_id_idx on agent_messages (task_id);
create index if not exists agent_messages_from_agent_idx on agent_messages (from_agent_id, created_at desc);
