-- MUDRO 004_post_comments.sql

create table if not exists post_comments (
  id bigserial primary key,
  post_id bigint not null references posts(id) on delete cascade,
  source text not null check (source in ('vk', 'tg')),
  source_comment_id text not null,
  source_parent_comment_id text,
  author_name text,
  published_at timestamptz not null,
  text text,
  reactions jsonb,
  media jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create unique index if not exists post_comments_source_source_comment_id_uq
  on post_comments (source, source_comment_id);

create index if not exists post_comments_post_id_published_at_idx
  on post_comments (post_id, published_at asc, id asc);
