-- MUDRO 006_media_assets.sql
-- Нормализованное хранилище медиа-ассетов и link-таблицы для постов и комментариев.

create table if not exists media_assets (
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
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  check (btrim(asset_key) <> ''),
  check (btrim(source) <> ''),
  check (btrim(kind) <> '')
);

create index if not exists media_assets_source_idx on media_assets (source);
create index if not exists media_assets_kind_idx on media_assets (kind);

create table if not exists post_media_links (
  post_id bigint not null references posts(id) on delete cascade,
  media_asset_id bigint not null references media_assets(id) on delete cascade,
  position integer not null default 1,
  created_at timestamptz not null default now(),
  primary key (post_id, position),
  check (position > 0)
);

create index if not exists post_media_links_media_idx on post_media_links (media_asset_id);

create table if not exists comment_media_links (
  comment_id bigint not null references post_comments(id) on delete cascade,
  media_asset_id bigint not null references media_assets(id) on delete cascade,
  position integer not null default 1,
  created_at timestamptz not null default now(),
  primary key (comment_id, position),
  check (position > 0)
);

create index if not exists comment_media_links_media_idx on comment_media_links (media_asset_id);
