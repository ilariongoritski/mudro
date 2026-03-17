---
name: postgres-optimization
description: Оптимизация PostgreSQL схемы, запросов и безопасности для проекта mudro
---

# Skill: PostgreSQL Optimization Patterns

Адаптирован для `mudro` (Postgres 15, таблицы: `posts`, `post_comments`, `media_assets`, `post_media_links`, `comment_reactions`, `agent_queue`).

## Индексы (текущая схема mudro)

### Критичные для ленты
```sql
-- posts: главный query path для API /api/front и /api/posts
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_source_published
    ON posts(source, published_at DESC);

-- cursor pagination (before_ts + before_id)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_cursor
    ON posts(published_at DESC, id DESC);

-- фильтрация по source_post_id (дедуп + импорт)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_source_post_id
    ON posts(source, source_post_id);
```

### Для комментариев
```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_post_comments_post_id
    ON post_comments(post_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_post_comments_parent
    ON post_comments(parent_comment_id) WHERE parent_comment_id IS NOT NULL;
```

### Для media
```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_post_media_links_post_id
    ON post_media_links(post_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_media_assets_kind
    ON media_assets(kind);
```

## Анализ запросов

```sql
-- Найти медленные запросы (нужен pg_stat_statements)
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- EXPLAIN для конкретного запроса
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT * FROM posts WHERE source = 'tg' ORDER BY published_at DESC LIMIT 20;
```

## Типичные оптимизации для mudro

### Cursor pagination (уже используется в API)
```sql
-- Правильно — использует индекс
SELECT * FROM posts
WHERE (published_at, id) < ($1, $2)
ORDER BY published_at DESC, id DESC
LIMIT $3;

-- Неправильно (OFFSET — медленно на больших таблицах)
SELECT * FROM posts ORDER BY published_at DESC LIMIT $1 OFFSET $2;
```

### JOIN с медиа (оптимальная выборка)
```sql
SELECT p.id, p.text, p.source,
       json_agg(ma.*) FILTER (WHERE ma.id IS NOT NULL) as media
FROM posts p
LEFT JOIN post_media_links pml ON pml.post_id = p.id
LEFT JOIN media_assets ma ON ma.id = pml.asset_id
WHERE p.source = 'tg'
GROUP BY p.id
ORDER BY p.published_at DESC
LIMIT 20;
```

## Мониторинг размера таблиц

```sql
SELECT schemaname, tablename,
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
       pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

## Безопасность (уже применено на VPS)
- Postgres доступен только через loopback на VPS
- Для приложения используется роль `mudro_app` (не суперпользователь)
- Внешний порт 5433 защищен systemd firewall guard

## Когда применять CONCURRENTLY
- Всегда при создании индексов в production (`CREATE INDEX CONCURRENTLY`)
- Не требует lock на таблицу
- Нельзя использовать внутри транзакции
