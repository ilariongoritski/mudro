# Movie Catalog Security Fixes

## Проблемы

### CRITICAL: Неправильные имена таблиц в importer

**Файл:** [`tools/importers/moviecatalogimport/cmd/main.go`](tools/importers/moviecatalogimport/cmd/main.go)

Используются имена без схемы (`movie_catalog_movies` вместо `movie_catalog.movies`):
- Строка 74: `insert into movie_catalog_movies`
- Строка 92: `delete from movie_catalog_movie_genres`
- Строка 105: `insert into movie_catalog_genres`
- Строка 114: `insert into movie_catalog_movie_genres`
- Строка 125: `delete from movie_catalog_movie_genres`
- Строка 128: `delete from movie_catalog_genres`
- Строка 131: `delete from movie_catalog_movies`
- Строка 136: `delete from movie_catalog_movies`

**Последствия:** Runtime ошибка если `search_path` не включает `movie_catalog`.

**Решение:** Заменить все на qualified имена:
- `movie_catalog_movies` → `movie_catalog.movies`
- `movie_catalog_genres` → `movie_catalog.genres`
- `movie_catalog_movie_genres` → `movie_catalog.movie_genres`

### WARNING: Неправильный default DSN

**Файл:** [`services/movie-catalog/app/run.go`](services/movie-catalog/app/run.go:22)

```go
defaultDSN = "postgres://postgres:postgres@db:5432/gallery?sslmode=disable&search_path=movie_catalog,public"
```

Указана БД `gallery` вместо `movie_catalog`.

**Решение:**
```go
defaultDSN = "postgres://postgres:postgres@db:5432/movie_catalog?sslmode=disable"
```

### WARNING: Потенциальная проблема с пустым array

**Файл:** [`services/movie-catalog/internal/adapters/postgrescatalog/repository.go`](services/movie-catalog/internal/adapters/postgrescatalog/repository.go:81)

Параметр `query.ExcludeGenres` передаётся как `$4::text[]`. Если это `nil` slice, pgx может обработать некорректно.

**Решение:** Явная инициализация в начале функций `ListMovies` и `countMovies`:
```go
excludeGenres := query.ExcludeGenres
if excludeGenres == nil {
    excludeGenres = []string{}
}
```

## План исправлений

1. Исправить все 8 SQL-запросов в importer
2. Исправить default DSN в app/run.go
3. Добавить защиту от nil slice в repository.go
4. Проверить что все запросы используют qualified имена

## Приоритет

CRITICAL — без этих фиксов importer не работает.
