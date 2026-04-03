-- MUDRO 011_drop_legacy_jsonb.sql
-- Удаление legacy jsonb-колонок после миграции данных в нормализованные таблицы.
--
-- Перед применением убедитесь, что все данные перенесены:
--   select count(*) from posts
--     where media is not null and media != 'null'::jsonb and media != '[]'::jsonb
--     and not exists (select 1 from post_media_links where post_id = posts.id);
--   -- Должно вернуть 0
--
--   select count(*) from post_comments
--     where media is not null and media != 'null'::jsonb and media != '[]'::jsonb
--     and not exists (select 1 from comment_media_links where comment_id = post_comments.id);
--   -- Должно вернуть 0

alter table posts drop column if exists media;
alter table post_comments drop column if exists media;
alter table post_comments drop column if exists reactions;
