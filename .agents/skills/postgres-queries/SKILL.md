---
name: postgres-queries
description: Шаблоны частых SQL-запросов для анализа данных проекта mudro
---

# Skill: Postgres Query Templates

## Подключение
Канонический DSN: `postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable`

Через Docker (по умолчанию в Makefile):
```bash
docker compose exec -T db psql -U postgres -d gallery
```

## Обзор данных

### Общая статистика
```sql
-- Посты по источникам
SELECT source, count(*) as posts FROM posts GROUP BY source ORDER BY posts DESC;

-- Общее количество
SELECT count(*) as total_posts FROM posts;
SELECT count(*) as total_comments FROM post_comments;
SELECT count(*) as total_media_assets FROM media_assets;
SELECT count(*) as total_comment_reactions FROM comment_reactions;
```

### Media coverage
```sql
-- Сколько постов имеют media
SELECT count(*) as posts_with_media
FROM posts
WHERE media IS NOT NULL AND media != '[]'::jsonb AND media != 'null'::jsonb;

-- Нормализованные media assets по типу
SELECT kind, count(*) FROM media_assets GROUP BY kind ORDER BY count DESC;

-- Посты без публичного media URL
SELECT count(*) FROM media_assets WHERE url IS NULL OR url = '';
```

### Комментарии
```sql
-- Распределение комментариев по постам (top-10 обсуждаемых)
SELECT p.source_post_id, p.source, count(c.id) as comments
FROM posts p
JOIN post_comments c ON c.post_id = p.id
GROUP BY p.id, p.source_post_id, p.source
ORDER BY comments DESC
LIMIT 10;

-- Комментарии с реакциями
SELECT count(DISTINCT cr.comment_id) as comments_with_reactions,
       count(*) as total_reactions
FROM comment_reactions cr;
```

### Проверка целостности
```sql
-- Orphan комментарии (без parent поста)
SELECT count(*) FROM post_comments WHERE post_id NOT IN (SELECT id FROM posts);

-- Дублирующиеся source_post_id
SELECT source, source_post_id, count(*)
FROM posts
GROUP BY source, source_post_id
HAVING count(*) > 1;

-- Media links без assets
SELECT count(*) FROM post_media_links WHERE asset_id NOT IN (SELECT id FROM media_assets);
```

### Временные рамки данных
```sql
-- Диапазон дат постов по источникам
SELECT source,
       min(published_at) as earliest,
       max(published_at) as latest,
       count(*) as total
FROM posts
GROUP BY source;
```

## Безопасность
- Через MCP сервер `mudro_postgres` — только **read-only** запросы
- **Никаких** `DROP`, `TRUNCATE`, `DELETE`, `UPDATE` без явного подтверждения
- Миграции применять только через `make migrate-*`
