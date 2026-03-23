# MUDROTOP

`MUDROTOP` теперь организован как staging-репозиторий в стиле `mini-mudro` для подготовки каталога фильмов к дальнейшему слиянию в основной `mudro`.

## Активные зоны

- `frontend/` - новый `Vite + React + TypeScript` UI
- `services/movie-catalog/` - read-only HTTP service на Go
- `contracts/http/movie-catalog-v1.yaml` - стабильный HTTP контракт
- `migrations/movie_catalog/` - схема staging `Postgres`
- `tools/importers/moviecatalogimport/` - импорт slim dataset в `Postgres`
- `scripts/prepare-movie-catalog-data.mjs` - нормализация raw dataset
- `ops/compose/docker-compose.local.yml` - локальный `Postgres` на `5434`
- `docs/` - доменные заметки и migration note

## Staging flow

1. Держать raw dataset вне репозитория:
   `D:\mudr\staging\mudrotop-source\raw\merged.json`
2. Подготовить slim dataset:
   `node scripts/prepare-movie-catalog-data.mjs`
3. Поднять локальный `Postgres`:
   `docker compose -f ops/compose/docker-compose.local.yml up -d`
4. Применить миграции из `migrations/movie_catalog/0001_init.sql`
   или `go run ./tools/migrations/moviecatalogmigrate/cmd`
5. Импортировать данные:
   `go run ./tools/importers/moviecatalogimport/cmd`
6. Запустить сервис:
   `go run ./services/movie-catalog/cmd`
7. Запустить frontend:
   `npm --prefix frontend install`
   `npm --prefix frontend run dev`

## Env

- service:
  - `MOVIE_CATALOG_ADDR`
  - `MOVIE_CATALOG_DB_DSN`
  - `MOVIE_CATALOG_IMPORT_FILE`
  - `MOVIE_CATALOG_MIGRATION_FILE`
- frontend:
  - `VITE_MOVIE_CATALOG_API_BASE_URL`

## Полный локальный compose

Если нужен не только `Postgres`, но и контейнерный запуск `service + frontend`:

`docker compose -f ops/compose/docker-compose.full.local.yml up --build`

## Важное

- raw `merged.json` не хранится в repo
- SQL не живёт в HTTP handlers
- frontend не фильтрует весь каталог локально
- старый `CRA`-контур больше не считается активной архитектурой
