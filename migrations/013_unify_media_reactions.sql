-- MUDRO 013_unify_media_reactions.sql
-- Унификация: все медиа через media_assets, реакции поддерживают картинки

-- ============================================================
-- 1. MEDIA_ASSETS — расширение kind для gif/sticker/custom_emoji
-- ============================================================

-- kind уже text с CHECK btrim <> '', но зафиксируем допустимые значения
-- в коде (Go enum), а не в БД — гибче для расширения.
-- Добавим поле duration для видео/gif:
alter table media_assets
  add column if not exists duration_ms integer,
  add column if not exists file_size_bytes bigint,
  add column if not exists blurhash text;

-- ============================================================
-- 2. CUSTOM EMOJI / STICKER PACKS — кастомные реакции
-- ============================================================

create table if not exists sticker_packs (
  id bigserial primary key,
  slug text not null unique,
  title text not null,
  creator_id bigint references users(id) on delete set null,
  is_public boolean not null default true,
  created_at timestamptz not null default now()
);

create table if not exists stickers (
  id bigserial primary key,
  pack_id bigint not null references sticker_packs(id) on delete cascade,
  media_asset_id bigint not null references media_assets(id) on delete cascade,
  emoji_alias text,
  position integer not null default 0,
  created_at timestamptz not null default now()
);

create index if not exists stickers_pack_id_idx on stickers (pack_id, position);
create index if not exists stickers_media_asset_id_idx on stickers (media_asset_id);

-- ============================================================
-- 3. РЕАКЦИИ — опциональная ссылка на media_asset (gif/sticker)
-- ============================================================

-- post_user_reactions: добавить media_asset_id для картинки-реакции
alter table post_user_reactions
  add column if not exists media_asset_id bigint references media_assets(id) on delete set null;

-- comment_user_reactions: то же
alter table comment_user_reactions
  add column if not exists media_asset_id bigint references media_assets(id) on delete set null;

-- message_reactions: то же
alter table message_reactions
  add column if not exists media_asset_id bigint references media_assets(id) on delete set null;

-- Агрегатные таблицы (post_reactions, comment_reactions) — тоже ссылка
-- чтобы при рендере агрегата показывать превью кастомной реакции
alter table post_reactions
  add column if not exists media_asset_id bigint references media_assets(id) on delete set null;

alter table comment_reactions
  add column if not exists media_asset_id bigint references media_assets(id) on delete set null;

-- ============================================================
-- 4. ПОМЕТИТЬ LEGACY JSONB КОЛОНКИ — для постепенной миграции
-- ============================================================

-- Не дропаем сразу — Go-код может читать оба источника,
-- но добавляем комментарии и nullable, чтоб новый код
-- использовал только *_media_links таблицы.

comment on column posts.media is 'LEGACY: use post_media_links instead';
comment on column post_comments.media is 'LEGACY: use comment_media_links instead';
comment on column post_comments.reactions is 'LEGACY: use comment_reactions + comment_user_reactions instead';
comment on column messages.media is 'LEGACY: use message_media_links instead';

-- ============================================================
-- 5. ИНДЕКСЫ для Go API запросов по медиа
-- ============================================================

-- "Все gif в посте" — kind фильтр
create index if not exists media_assets_kind_created_idx
  on media_assets (kind, created_at desc);

-- "Медиа юзера" — через posts.author_id + post_media_links
-- уже покрыто существующими индексами

-- Быстрый поиск стикера по emoji_alias
create index if not exists stickers_emoji_alias_idx
  on stickers (emoji_alias) where emoji_alias is not null;
