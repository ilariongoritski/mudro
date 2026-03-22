# MUDRO - microservices-first монорепа (Go + Postgres + React)

MUDRO - это монорепозиторий с активным runtime-контуром в `services/*`, операционными CLI в `tools/*`, и архивной зоной в `legacy/old/*`.

## Что сейчас считается каноном

- Активные runtime-сервисы:
- `services/feed-api`
- `services/agent`
- `services/bot`
- Legacy-контур (не используется по умолчанию):
- `legacy/old/services/reporter-old`
- Runtime-папки `cmd/api|agent|bot|reporter` выведены в `legacy/old/cmd-runtime/*`.
- `cmd/*` в активной зоне - это compatibility forwarding для CLI в `tools/*`.

Подробная структура:
- [docs/service-catalog.md](docs/service-catalog.md)
- [docs/repository-topology.md](docs/repository-topology.md)
- [docs/repo-layout.md](docs/repo-layout.md)

## Требования

- Docker + Docker Compose
- Go 1.22+
- Windows: WSL2 + Docker Desktop (WSL integration)

## Быстрый старт (локально)

```bash
make up
make dbcheck
make migrate
make tables
go test ./...
```

Канонический DSN для локалки:
`postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable`

## Структура репозитория

- `services/` - runtime-сервисы
- `tools/` - importers/backfill/maintenance CLI
- `internal/` - общая доменная и прикладная логика
- `contracts/` - контракты HTTP и событий
- `ops/` - compose, runbooks, ops-скрипты, env-описания
- `platform/agent-control/` - правила, политики, профили LLM-агентов
- `legacy/old/` - архивная зона (Old)
- `migrations/` - SQL-миграции
- `frontend/` - React + TypeScript

## Канонические точки запуска

```bash
go run ./services/feed-api/cmd
go run ./services/agent/cmd
go run ./services/bot/cmd
```

Legacy reporter (только при явной необходимости):

```bash
go run ./legacy/old/services/reporter-old/cmd
# или
make legacy-report-run
```

## Makefile: основные цели

- `make health` - базовый health loop
- `make bot-run` - запуск Telegram bot
- `make agent-plan-once` - единичный проход planner
- `make agent-plan` - planner в цикле
- `make agent-work` - worker в цикле
- `make agent-approve TASK_ID=<id>`
- `make agent-reject TASK_ID=<id> REASON='...'`
- `make media-backfill`
- `make comment-backfill`
- `make tg-csv-import CSV=/abs/path/file.csv`
- `make tg-comments-csv-import CSV=/abs/path/file.csv`
- `make tg-comment-media-import DIR=/abs/path/dir`
- `make memento`

## CLI-контур (`tools/*`)

### Importers

- `tools/importers/vkimport`
- `tools/importers/tgimport`
- `tools/importers/tgload`
- `tools/importers/tgcsvimport`
- `tools/importers/tgcommentsimport`
- `tools/importers/tgcommentscsvimport`
- `tools/importers/tgcommentmediaimport`
- `tools/importers/tghtmlimport`

Примеры:

```bash
go run ./tools/importers/vkimport/cmd -dir ~/vk-export -dsn "postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable"
go run ./tools/importers/tgimport/cmd -in result.json -out feed_items.json
go run ./tools/importers/tgload/cmd -in feed_items.json -dsn "postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable"
```

### Backfill

- `tools/backfill/mediabackfill`
- `tools/backfill/commentbackfill`

### Maintenance

- `tools/maintenance/memento`
- `tools/maintenance/tgdedupe`
- `tools/maintenance/tgrootmerge`

## Frontend

Папка: `frontend/`

```bash
cd frontend
npm.cmd install
npm.cmd run dev
npm.cmd run build
npm.cmd run lint
```

## Compose-профили

- Core: `ops/compose/docker-compose.core.yml`
- Legacy: `ops/compose/docker-compose.legacy.yml`

Проверка конфигурации:

```bash
docker compose -f ops/compose/docker-compose.core.yml config
docker compose -f ops/compose/docker-compose.legacy.yml config
```

## Контракты

- HTTP: `contracts/http/feed-api.yaml`
- Events: `contracts/events/agent-task-events.yaml`

## Agent governance

Gateway-файл: `AGENTS.md`

Каноничные правила и профили:
- `platform/agent-control/AGENTS.core.md`
- `platform/agent-control/policies/*`
- `platform/agent-control/profiles/*`
- `platform/agent-control/services-map.yaml`

## Операционные документы

- Runbook: `ops/runbooks/ops-runbook.md`
- Worker autonomy: `docs/worker-autonomy.md`
- Microservices architecture: `docs/microservices-architecture.md`

## Примечания

- `VK` в проекте считается snapshot-only источником.
- Публичный API-контракт и миграции БД в этом этапе не ломались; менялась структура репозитория и контуры запуска.
- Все устаревшее и неиспользуемое должно уезжать в `legacy/old/*` с фиксацией в `legacy/old/manifest.yaml`.

## Runtime Bootstrap (P0)

Use these commands as the canonical local bootstrap path:

```bash
make core-up
make dbcheck-core
make migrate-runtime
make tables-core
make test-active
make count-posts-core
```

Full health loop:

```bash
make health-runtime
```

Backward compatibility:

```bash
make health
```

## Local Demo (No Vercel)

Terminal 1:

```bash
make demo-up
```

`make demo-up` now also seeds the local demo feed from `data/nu/feed_items.json` when `posts` is still empty.

Terminal 2:

```bash
npm.cmd --prefix frontend run dev
```

Check:

```bash
make demo-check
```

Open:
- `http://127.0.0.1:5173`
- `http://127.0.0.1:8080/healthz`
