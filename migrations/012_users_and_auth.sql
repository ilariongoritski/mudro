-- 012: users table, post_user_likes with trigger, extend comment source

create table if not exists users (
  id         bigserial    primary key,
  username   text         not null unique,
  email      text         not null unique,
  password_hash text      not null,
  display_name  text,
  avatar_url    text,
  created_at timestamptz  not null default now(),
  updated_at timestamptz  not null default now()
);

create index if not exists users_username_idx on users (username);
create index if not exists users_email_idx    on users (email);

-- local user likes (separate from external-platform post_account_likes)
create table if not exists post_user_likes (
  post_id    bigint      not null references posts (id) on delete cascade,
  user_id    bigint      not null references users (id) on delete cascade,
  created_at timestamptz not null default now(),
  primary key (post_id, user_id)
);

create or replace function update_post_user_likes_count() returns trigger as $$
begin
  if tg_op = 'INSERT' then
    update posts set likes_count = likes_count + 1, updated_at = now()
      where id = new.post_id;
  elsif tg_op = 'DELETE' then
    update posts set likes_count = greatest(likes_count - 1, 0), updated_at = now()
      where id = old.post_id;
  end if;
  return null;
end;
$$ language plpgsql;

drop trigger if exists trg_post_user_likes_count on post_user_likes;
create trigger trg_post_user_likes_count
  after insert or delete on post_user_likes
  for each row execute function update_post_user_likes_count();

-- allow local comments
alter table post_comments drop constraint if exists post_comments_source_check;
alter table post_comments add constraint post_comments_source_check
  check (source in ('vk', 'tg', 'local'));
