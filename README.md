# MUDRO

MUDRO — Go-first monorepo для пред-MVP социального продукта: лента VK/TG-контента, авторизация, чат, casino showcase, agent-worker контур, Telegram-боты, import/backfill инструменты и self-hosted/VPS runtime.

Текущее состояние: `v0.1.0-mvp`, pre-MVP launch readiness. Код собирается и тестируется, но публичный запуск требует внешней ротации секретов, настройки VPS/HTTPS и применения миграций на целевых БД.

## MVP Состояние

Готово для инженерной проверки:

- backend `feed-api` с `/healthz`, `/api/front`, auth/admin/chat/casino proxy;
- frontend React/Vite с лентой, login/register, chat, casino, admin guard;
- отдельный casino runtime с собственной БД и internal secret между `feed-api` и `casino-api`;
- up-only миграционный контур с guard против случайного применения `*.down.sql`;
- CI workflow для Go/frontend/contracts/migration safety/integration DB/release smoke;
- production compose с обязательными секретами и non-superuser DSN;
- nginx/HTTPS/secret rotation runbook-и для VPS.

Не считать готовым без ручной проверки:

- реальные секреты должны быть ротированы вне репозитория;
- роли `mudro_app` и `mudro_casino_app` должны быть созданы/grant на VPS БД;
- HTTPS и домен должны быть подняты по runbook;
- локальная браузерная демка должна пройти smoke по `/`, `/login`, `/casino`, `/chat`, `/admin`.

## Быстрый Локальный Запуск

Основной локальный контур:

```bash
make core-up
make dbcheck-core
make migrate-runtime
make tables-core
make health-runtime
```

Отдельный локальный casino contour:

```bash
make casino-up
make health-casino
```

Frontend:

```bash
cd frontend
npm.cmd install
npm.cmd run dev
```

Ожидаемые адреса:

- Frontend: `http://127.0.0.1:5173`
- Feed API: `http://127.0.0.1:8080/healthz`
- Casino API: `http://127.0.0.1:8082/healthz`

Проверка демо:

```bash
make demo-check
```

Если БД пустая, нужен seed/import данных. Для полноценного локального просмотра сначала убедиться, что `/api/front?limit=1` отдаёт непустую ленту.

## Архитектура

Активные runtime-сервисы:

- `services/feed-api/cmd` — основной HTTP API для MVP: feed, auth, admin, chat, casino proxy.
- `services/casino/cmd/casino` — отдельный casino service с собственной Postgres БД.
- `services/agent/cmd` — planner/worker automation.
- `services/bot/cmd` — Telegram control-plane.

Additive/bootstrap сервисы:

- `services/api-gateway/cmd` — будущий edge/API gateway.
- `services/bff-web/cmd` — web aggregation layer.
- `services/auth-api/cmd` — выделяемая auth/admin граница.
- `services/orchestration-api/cmd` — выделяемая orchestration/status граница.

Инструменты:

- `tools/importers/*` — импорт VK/TG и CSV/HTML источников.
- `tools/backfill/*` — backfill медиа/комментариев.
- `tools/maintenance/*` — dedupe, merge, диагностика.
- `tools/validate-contracts` — проверка HTTP contracts.

## Compose Контуры

- `ops/compose/docker-compose.core.yml` — canonical local runtime: db, redis, kafka, api, agent, bff/movie-catalog.
- `ops/compose/docker-compose.casino.local.yml` — локальный casino DB/API без prod secrets.
- `ops/compose/docker-compose.services.yml` — дополнительные сервисы для микросервисной декомпозиции.
- `docker-compose.prod.yml` — production-style release stack для VPS/self-hosted.
- `docker-compose.yml` — legacy/deprecated compatibility файл.

Production compose намеренно fail-fast: без `MUDRO_APP_DSN`, `CASINO_APP_DSN`, `JWT_SECRET`, MinIO credentials и `CASINO_INTERNAL_SECRET` он не должен стартовать. Telegram-боты запускаются отдельным контуром, чтобы не расширять blast radius основного MVP runtime.

## Миграции

Основная БД:

```bash
make migrate-list
make migrate-all-dry-run
make migrate-all
```

Casino БД:

```bash
make migrate-casino-list
make migrate-casino-dry-run
make migrate-casino
```

Guard:

```bash
make check-migration-up-list
```

Обычный bootstrap/recovery не должен применять `*.down.sql`.

## Проверки

Базовый набор:

```bash
make selftest
go test ./...
go build ./services/... ./tools/... ./cmd/... ./pkg/...
go run ./tools/validate-contracts -dir contracts
cd frontend && npm.cmd run lint
cd frontend && npm.cmd run test
cd frontend && npm.cmd run build
docker compose -f ops/compose/docker-compose.core.yml config
docker compose -f ops/compose/docker-compose.casino.local.yml config
```

Production config проверять только с dummy env или в закрытом CI/VPS shell. Не публиковать вывод `docker compose config`, он раскрывает значения env:

```bash
docker compose -f docker-compose.prod.yml config
```

## VPS / Release

Каноничный MVP target: self-hosted/VPS через `docker-compose.prod.yml` + nginx/HTTPS.

Документы:

- `ops/runbooks/ops-runbook.md`
- `ops/runbooks/vps-https-nginx.md`
- `ops/runbooks/secrets-rotation.md`
- `docs/release-showcase-checklist.md`

Перед запуском:

- ротировать Telegram/OpenAI/OpenRouter/JWT/MinIO secrets;
- создать non-superuser DB roles;
- заполнить env на VPS вне git;
- применить up-only миграции;
- пройти smoke: `/healthz`, `/api/front`, frontend route refresh, casino balance/action/history.

## GitHub / CI Состояние

В репозитории есть GitHub Actions workflow `.github/workflows/ci.yml`.

CI покрывает:

- Go build/test;
- migration safety;
- backend unit/integration с Postgres service;
- e2e smoke с явным DSN;
- frontend lint/test/build;
- HTTP contracts validation;
- release compose и Docker build smoke.

Текущие локальные изменения могут быть не закоммичены. Перед PR/merge обязательно проверить `git status --short`, не смешивать unrelated WIP и не коммитить реальные секреты.

## Legacy

Legacy зоны не являются основным runtime:

- `legacy/old/`
- root `docker-compose.yml`
- исторические `cmd/*` stubs

Использовать их только для recovery/compatibility задач.
