# movie-catalog

Read-only staging service для подготовки `MudroTop` к будущему merge в `mudro`.

## Env

- `MOVIE_CATALOG_ADDR` — HTTP address, по умолчанию `:8091`
- `MOVIE_CATALOG_DB_DSN` — Postgres DSN, по умолчанию `postgres://postgres:postgres@localhost:5434/movie_catalog?sslmode=disable`
- `MOVIE_CATALOG_IMPORT_FILE` — путь для importer, по умолчанию `out/movie-catalog.slim.json`

## Контур данных

1. `node scripts/prepare-movie-catalog-data.mjs`
2. `go run ./tools/migrations/moviecatalogmigrate/cmd`
3. `go run ./tools/importers/moviecatalogimport/cmd`
4. `go run ./services/movie-catalog/cmd`

## API

- `GET /healthz`
- `GET /api/movie-catalog/genres`
- `GET /api/movie-catalog/movies`
