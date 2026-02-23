-- MUDRO 001_init.sql

create table if not exists posts (
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

create unique index if not exists posts_source_source_post_id_uq
  on posts (source, source_post_id);

create index if not exists posts_published_at_id_idx
  on posts (published_at desc, id desc);

create table if not exists post_reactions (
  post_id bigint not null references posts(id) on delete cascade,
  emoji text not null,
  count int not null default 0,
  primary key (post_id, emoji)
);