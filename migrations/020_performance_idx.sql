-- MUDRO 020_performance_idx.sql
-- Оптимизация поиска через триграммы и GIN индексы.

create extension if not exists pg_trgm;

-- Индекс для быстрого поиска по тексту постов (LOWER(text) LIKE %...%)
create index if not exists idx_posts_text_trgm on posts using gin (lower(text) gin_trgm_ops);

-- Индекс для быстрого поиска по именам авторов комментариев
create index if not exists idx_post_comments_author_trgm on post_comments using gin (lower(author_name) gin_trgm_ops);

-- Анализ таблиц после создания индексов
analyze posts;
analyze post_comments;
