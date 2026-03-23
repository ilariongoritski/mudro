# Movie Catalog migrations

Первая волна использует одну базовую миграцию:

- `0001_init.sql`

Она создаёт:

- `movie_catalog_movies`
- `movie_catalog_genres`
- `movie_catalog_movie_genres`

Индексная стратегия уже включена в эту миграцию и рассчитана на read-only фильтрацию:

- по `year`
- по `duration_minutes`
- по связующей таблице жанров
