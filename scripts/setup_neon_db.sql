-- MUDRO: Unified Neon DB Setup (Migrations + MVP Seeds)
-- Copy and paste this into Neon SQL Editor

-- 1. Extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 2. Core Tables
CREATE TABLE IF NOT EXISTS posts (
  id bigserial primary key,
  source text not null check (source in ('vk','tg')),
  source_post_id text not null,
  published_at timestamptz not null,
  text text,
  media jsonb,
  likes_count int not null default 0,
  views_count int,
  comments_count int,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);
CREATE UNIQUE INDEX IF NOT EXISTS posts_source_source_post_id_uq ON posts (source, source_post_id);
CREATE INDEX IF NOT EXISTS posts_published_at_id_idx ON posts (published_at desc, id desc);

CREATE TABLE IF NOT EXISTS post_reactions (
  post_id bigint not null references posts(id) on delete cascade,
  emoji text not null,
  count int not null default 0,
  media_asset_id bigint, -- will be linked later
  primary key (post_id, emoji)
);

CREATE TABLE IF NOT EXISTS accounts (
  id bigserial primary key,
  external_id text not null,
  platform text not null default 'local',
  display_name text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique (platform, external_id)
);
CREATE INDEX IF NOT EXISTS accounts_platform_external_id_idx ON accounts (platform, external_id);

CREATE TABLE IF NOT EXISTS post_account_likes (
  post_id bigint not null references posts(id) on delete cascade,
  account_id bigint not null references accounts(id) on delete cascade,
  created_at timestamptz not null default now(),
  primary key (post_id, account_id)
);

CREATE TABLE IF NOT EXISTS post_comments (
  id bigserial primary key,
  post_id bigint not null references posts(id) on delete cascade,
  source text not null check (source in ('vk', 'tg')),
  source_comment_id text not null,
  source_parent_comment_id text,
  parent_comment_id bigint references post_comments(id) on delete set null,
  author_name text,
  published_at timestamptz not null,
  text text,
  reactions jsonb,
  media jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);
CREATE UNIQUE INDEX IF NOT EXISTS post_comments_source_source_comment_id_uq ON post_comments (source, source_comment_id);
CREATE INDEX IF NOT EXISTS post_comments_post_id_published_at_idx ON post_comments (post_id, published_at asc, id asc);
CREATE INDEX IF NOT EXISTS post_comments_parent_comment_id_idx ON post_comments (parent_comment_id);

CREATE TABLE IF NOT EXISTS comment_reactions (
  comment_id bigint not null references post_comments(id) on delete cascade,
  emoji text not null,
  count integer not null default 0,
  media_asset_id bigint, -- will be linked later
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  primary key (comment_id, emoji),
  check (btrim(emoji) <> ''),
  check (count >= 0)
);

-- 3. Media Assets
CREATE TABLE IF NOT EXISTS media_assets (
  id bigserial primary key,
  asset_key text not null unique,
  source text not null,
  kind text not null,
  original_url text,
  preview_url text,
  title text,
  mime_type text,
  width integer,
  height integer,
  extra jsonb not null default '{}'::jsonb,
  duration_ms integer,
  file_size_bytes bigint,
  blurhash text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  check (btrim(asset_key) <> ''),
  check (btrim(source) <> ''),
  check (btrim(kind) <> '')
);

CREATE TABLE IF NOT EXISTS post_media_links (
  post_id bigint not null references posts(id) on delete cascade,
  media_asset_id bigint not null references media_assets(id) on delete cascade,
  position integer not null default 1,
  created_at timestamptz not null default now(),
  primary key (post_id, position),
  check (position > 0)
);

CREATE TABLE IF NOT EXISTS comment_media_links (
  comment_id bigint not null references post_comments(id) on delete cascade,
  media_asset_id bigint not null references media_assets(id) on delete cascade,
  position integer not null default 1,
  created_at timestamptz not null default now(),
  primary key (comment_id, position),
  check (position > 0)
);

-- 4. Users & Auth
CREATE TABLE IF NOT EXISTS users (
  id bigserial primary key,
  email text not null unique,
  telegram_id bigint,
  telegram_username text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);
CREATE UNIQUE INDEX IF NOT EXISTS users_telegram_id_uq ON users (telegram_id) WHERE telegram_id is not null;

CREATE TABLE IF NOT EXISTS auth_tokens (
  token text primary key,
  email text not null,
  expires_at timestamptz not null,
  created_at timestamptz not null default now()
);

CREATE TABLE IF NOT EXISTS user_sessions (
  id text primary key,
  user_id bigint not null references users(id) on delete cascade,
  expires_at timestamptz not null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

-- 5. Agents & Tasks
CREATE TABLE IF NOT EXISTS agent_queue (
  id bigserial primary key,
  kind text not null,
  payload jsonb not null default '{}'::jsonb,
  status text not null default 'queued',
  priority int not null default 0,
  attempts int not null default 0,
  max_attempts int not null default 3,
  dedupe_key text,
  run_after timestamptz not null default now(),
  locked_by text,
  locked_at timestamptz,
  last_error text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  finished_at timestamptz
);

CREATE TABLE IF NOT EXISTS agent_task_events (
  id bigserial primary key,
  event_id text not null unique,
  task_id bigint not null,
  event_type text not null,
  status text not null,
  kind text,
  dedupe_key text,
  error text,
  occurred_at timestamptz not null default now(),
  created_at timestamptz not null default now()
);

-- 6. Chat & Casino
CREATE TABLE IF NOT EXISTS chat_messages (
  id BIGSERIAL PRIMARY KEY,
  room_name TEXT NOT NULL,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  username TEXT NOT NULL,
  user_role TEXT NOT NULL DEFAULT 'user',
  body TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS casino_accounts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    type        TEXT NOT NULL DEFAULT 'user' CHECK (type IN ('user','system')),
    code        TEXT UNIQUE NOT NULL,
    currency    TEXT NOT NULL DEFAULT 'МДР',
    balance     NUMERIC(30,10) NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS casino_transfers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind        TEXT NOT NULL CHECK (kind IN ('bet_stake','bet_payout','deposit','withdrawal','adjustment')),
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS casino_ledger_entries (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transfer_id UUID NOT NULL REFERENCES casino_transfers(id) ON DELETE CASCADE,
    account_id  UUID NOT NULL REFERENCES casino_accounts(id) ON DELETE RESTRICT,
    direction   TEXT NOT NULL CHECK (direction IN ('debit','credit')),
    amount      NUMERIC(30,10) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS casino_rounds (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    server_seed      TEXT NOT NULL,
    server_seed_hash TEXT NOT NULL,
    client_seed      TEXT,
    nonce            INT NOT NULL DEFAULT 0,
    round_hash       TEXT,
    roll             INT,
    bet_amount       NUMERIC(30,10),
    payout_amount    NUMERIC(30,10),
    multiplier       NUMERIC(10,4),
    tier_label       TEXT,
    tier_symbol      TEXT,
    status           TEXT NOT NULL DEFAULT 'prepared' CHECK (status IN ('prepared','resolved','cancelled')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at      TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS casino_rtp_profiles (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT UNIQUE NOT NULL,
    rtp        NUMERIC(5,2) NOT NULL,
    paytable   JSONB NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 7. Triggers
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

-- 8. MVP Seeds
INSERT INTO casino_rtp_profiles (name, rtp, paytable, is_default) VALUES
('default', 96, '[{"minRoll":0,"maxRoll":0,"multiplier":25,"label":"МЕГА ДЖЕКПОТ","symbol":"🎰🎰🎰"},{"minRoll":1,"maxRoll":2,"multiplier":8,"label":"ДЖЕКПОТ","symbol":"💎💎💎"},{"minRoll":3,"maxRoll":7,"multiplier":3,"label":"СУПЕР","symbol":"🔔🔔🔔"},{"minRoll":8,"maxRoll":17,"multiplier":1.5,"label":"КРУТО","symbol":"🍒🍒🍒"},{"minRoll":18,"maxRoll":37,"multiplier":0.8,"label":"ХОРОШО","symbol":"🍋🍋🍋"},{"minRoll":38,"maxRoll":52,"multiplier":0.6,"label":"МЕЛОЧЬ","symbol":"🍇🍇🍇"},{"minRoll":53,"maxRoll":99,"multiplier":0,"label":"МИМО","symbol":"💩"}]'::jsonb, true),
('vip', 97, '[{"minRoll":0,"maxRoll":0,"multiplier":30,"label":"МЕГА ДЖЕКПОТ","symbol":"🎰🎰🎰"},{"minRoll":1,"maxRoll":2,"multiplier":8,"label":"ДЖЕКПОТ","symbol":"💎💎💎"},{"minRoll":3,"maxRoll":7,"multiplier":3,"label":"СУПЕР","symbol":"🔔🔔🔔"},{"minRoll":8,"maxRoll":17,"multiplier":1.5,"label":"КРУТО","symbol":"🍒🍒🍒"},{"minRoll":18,"maxRoll":37,"multiplier":0.8,"label":"ХОРОШО","symbol":"🍋🍋🍋"},{"minRoll":38,"maxRoll":52,"multiplier":0.333,"label":"МЕЛОЧЬ","symbol":"🍇🍇🍇"},{"minRoll":53,"maxRoll":99,"multiplier":0,"label":"МИМО","symbol":"💩"}]'::jsonb, false),
('shark', 94, '[{"minRoll":0,"maxRoll":0,"multiplier":20,"label":"МЕГА ДЖЕКПОТ","symbol":"🎰🎰🎰"},{"minRoll":1,"maxRoll":2,"multiplier":8,"label":"ДЖЕКПОТ","symbol":"💎💎💎"},{"minRoll":3,"maxRoll":7,"multiplier":3,"label":"СУПЕР","symbol":"🔔🔔🔔"},{"minRoll":8,"maxRoll":17,"multiplier":1.5,"label":"КРУТО","symbol":"🍒🍒🍒"},{"minRoll":18,"maxRoll":37,"multiplier":0.8,"label":"ХОРОШО","symbol":"🍋🍋🍋"},{"minRoll":38,"maxRoll":57,"multiplier":0.6,"label":"МЕЛОЧЬ","symbol":"🍇🍇🍇"},{"minRoll":58,"maxRoll":99,"multiplier":0,"label":"МИМО","symbol":"💩"}]'::jsonb, false)
ON CONFLICT (name) DO NOTHING;

INSERT INTO casino_accounts (type, code, currency, balance) VALUES
('system', 'SYSTEM_HOUSE_POOL', 'МДР', 1000000),
('system', 'SYSTEM_SETTLEMENT_POOL', 'МДР', 1000000)
ON CONFLICT (code) DO NOTHING;

INSERT INTO posts (source, source_post_id, published_at, text, media, likes_count, views_count, comments_count, updated_at) VALUES
('tg', 'promo-1', NOW() - INTERVAL '1 hour', 'Добро пожаловать в Mudro — платформу нового поколения! 🔥\nМы создали этот MVP, чтобы показать, как мощно может выглядеть ваша личная лента. Наслаждайтесь Glassmorphism-дизайном и плавными анимациями.', 
'[{"kind": "photo", "url": "https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?q=80&w=2000&auto=format&fit=crop", "position": 1, "width": 2000, "height": 1333}]', 124, 1500, 12, NOW()),
('tg', 'promo-2', NOW() - INTERVAL '3 hours', 'Темная тема (Hotpink Dark) создает по-настоящему премиальный опыт. Каждая карточка поста имеет эффект матового стекла (backdrop-filter) и мягкую неоновую подсветку при наведении. ✨', 
'[{"kind": "photo", "url": "https://images.unsplash.com/photo-1550684848-fac1c5b4e853?q=80&w=2000&auto=format&fit=crop", "position": 1, "width": 2000, "height": 1333}]', 89, 840, 5, NOW()),
('vk', 'promo-3', NOW() - INTERVAL '5 hours', 'База данных PostgreSQL успешно развернута, API-шлюз на Nginx работает без 404 ошибок, а фронтенд на Vercel безупречно проксирует все запросы. Это полностью рабочий Fullstack MVP! 🚀', 
'[{"kind": "photo", "url": "https://images.unsplash.com/photo-1550745165-9bc0b252726f?q=80&w=2000&auto=format&fit=crop", "position": 1, "width": 2000, "height": 1333}]', 256, 3100, 42, NOW())
ON CONFLICT (source, source_post_id) DO NOTHING;

INSERT INTO post_reactions (post_id, emoji, count)
SELECT id, '❤️', likes_count FROM posts WHERE source_post_id IN ('promo-1', 'promo-2', 'promo-3')
ON CONFLICT DO NOTHING;

DO $$ 
DECLARE
    post1_id bigint;
    post3_id bigint;
    root1_id bigint;
BEGIN
    SELECT id INTO post1_id FROM posts WHERE source_post_id = 'promo-1';
    SELECT id INTO post3_id FROM posts WHERE source_post_id = 'promo-3';

    IF post1_id IS NOT NULL THEN
        INSERT INTO post_comments (post_id, source, source_comment_id, author_name, published_at, text, reactions)
        VALUES (post1_id, 'tg', 'tg-c1', 'Alex Cyber', NOW() - INTERVAL '40 minutes', 'Вау! Выглядит потрясающе, особенно dark theme. Это именно то, что я ждал от MUDRO.', '[{"label":"🔥", "count": 15}]'::jsonb)
        ON CONFLICT (source, source_comment_id) DO NOTHING RETURNING id INTO root1_id;

        IF root1_id IS NULL THEN
            SELECT id INTO root1_id FROM post_comments WHERE source_comment_id = 'tg-c1';
        END IF;

        IF root1_id IS NOT NULL THEN
            INSERT INTO post_comments (post_id, source, source_comment_id, parent_comment_id, author_name, published_at, text, reactions)
            VALUES (post1_id, 'tg', 'tg-c1-r1', root1_id, 'MUDRO Team', NOW() - INTERVAL '30 minutes', 'Спасибо, Alex! Мы старались сделать фокус на UX и плавности анимации.', '[{"label":"❤️", "count": 5}]'::jsonb)
            ON CONFLICT (source, source_comment_id) DO NOTHING;
        END IF;

        INSERT INTO post_comments (post_id, source, source_comment_id, author_name, published_at, text)
        VALUES (post1_id, 'tg', 'tg-c2', 'Tech Reviewer', NOW() - INTERVAL '15 minutes', 'Блюр и glassmorphism отлично смотрятся на десктопе. Главное чтобы на мобилках не лагало.')
        ON CONFLICT (source, source_comment_id) DO NOTHING;
    END IF;

    IF post3_id IS NOT NULL THEN
        INSERT INTO post_comments (post_id, source, source_comment_id, author_name, published_at, text, reactions)
        VALUES (post3_id, 'vk', 'vk-c1', 'Dmitry DevOps', NOW() - INTERVAL '2 hours', 'Отлично, что API и база работают прямо из коробки на MVP. Архитектура solid. 👍', '[{"raw":"👍", "count": 22}]'::jsonb)
        ON CONFLICT (source, source_comment_id) DO NOTHING;
    END IF;
END $$;
