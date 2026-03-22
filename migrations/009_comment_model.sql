alter table post_comments
  add column if not exists parent_comment_id bigint references post_comments(id) on delete set null;

create index if not exists post_comments_parent_comment_id_idx
  on post_comments (parent_comment_id);

create table if not exists comment_reactions (
  comment_id bigint not null references post_comments(id) on delete cascade,
  emoji text not null,
  count integer not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  primary key (comment_id, emoji),
  check (btrim(emoji) <> ''),
  check (count >= 0)
);

create index if not exists comment_reactions_emoji_idx
  on comment_reactions (emoji);
