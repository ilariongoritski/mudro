# Movie Catalog

`movie-catalog` - staging-подсистема для подготовки `MudroTop` к будущему слиянию в основной `mudro`.

## Контур первой волны

- `services/movie-catalog` - read-only HTTP service на Go
- `contracts/http/movie-catalog-v1.yaml` - стабильный HTTP контракт
- `migrations/movie_catalog` - схема staging Postgres
- `tools/importers/moviecatalogimport` - idempotent import slim dataset в Postgres
- `scripts/prepare-movie-catalog-data.mjs` - подготовка slim JSON из raw `merged.json`
- `frontend/` - новый Vite + TypeScript UI

## Staging storage

- raw source: `D:\mudr\staging\mudrotop-source\raw\merged.json`
- slim output: `out/movie-catalog.slim.json`
- staging postgres: `postgres://postgres:postgres@localhost:5434/movie_catalog?sslmode=disable`

## Запрещённые практики

- SQL в HTTP handlers
- чтение raw `merged.json` в runtime request path
- giant JSON в `frontend/src`
- клиентская фильтрация всей базы
- промежуточная SQLite
