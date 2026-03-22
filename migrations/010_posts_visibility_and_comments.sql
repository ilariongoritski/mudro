-- MUDRO 010_posts_visibility_and_comments.sql
-- 1) Бэкфилл comments_count реальными значениями.
-- 2) Колонка visible (замена файлового TG-фильтра).
-- 3) Триггер автообновления comments_count при insert/delete комментариев.

-- Бэкфилл comments_count
update posts set comments_count = sub.cnt
from (
  select post_id, count(*) as cnt
  from post_comments
  group by post_id
) sub
where posts.id = sub.post_id
  and (posts.comments_count is null or posts.comments_count != sub.cnt);

update posts set comments_count = 0
where comments_count is null;

alter table posts
  alter column comments_count set not null,
  alter column comments_count set default 0;

-- Колонка видимости
alter table posts
  add column if not exists visible boolean not null default true;

create index if not exists posts_visible_published_at_idx
  on posts (published_at desc, id desc)
  where visible = true;

-- Триггер: автообновление comments_count
create or replace function update_post_comments_count() returns trigger as $$
begin
  if tg_op = 'INSERT' then
    update posts set comments_count = comments_count + 1, updated_at = now()
    where id = new.post_id;
  elsif tg_op = 'DELETE' then
    update posts set comments_count = greatest(comments_count - 1, 0), updated_at = now()
    where id = old.post_id;
  end if;
  return null;
end;
$$ language plpgsql;

drop trigger if exists trg_post_comments_count on post_comments;
create trigger trg_post_comments_count
  after insert or delete on post_comments
  for each row execute function update_post_comments_count();
