-- MUDRO 011_drop_legacy_jsonb.sql
-- Удаление legacy jsonb-колонок после миграции данных в нормализованные таблицы.

do $$
declare
  missing_post_links bigint := 0;
  missing_comment_links bigint := 0;
begin
  if exists (
    select 1
    from information_schema.columns
    where table_schema = current_schema()
      and table_name = 'posts'
      and column_name = 'media'
  ) then
    select count(*)
      into missing_post_links
      from posts
      where media is not null
        and media != 'null'::jsonb
        and media != '[]'::jsonb
        and not exists (
          select 1
          from post_media_links
          where post_media_links.post_id = posts.id
        );
  end if;

  if exists (
    select 1
    from information_schema.columns
    where table_schema = current_schema()
      and table_name = 'post_comments'
      and column_name = 'media'
  ) then
    select count(*)
      into missing_comment_links
      from post_comments
      where media is not null
        and media != 'null'::jsonb
        and media != '[]'::jsonb
        and not exists (
          select 1
          from comment_media_links
          where comment_media_links.comment_id = post_comments.id
        );
  end if;

  if missing_post_links > 0 then
    raise exception 'legacy posts.media still has % rows without post_media_links', missing_post_links;
  end if;

  if missing_comment_links > 0 then
    raise exception 'legacy post_comments.media still has % rows without comment_media_links', missing_comment_links;
  end if;
end $$;

alter table posts drop column if exists media;
alter table post_comments drop column if exists media;
alter table post_comments drop column if exists reactions;
