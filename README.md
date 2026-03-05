# MUDRO — Go backend for a unified VK + Telegram feed

MUDRO is a Go + Postgres backend with importers for VK export data and Telegram export JSON. It also includes a Telegram bot for operational summaries.

## Requirements
- Docker + Docker Compose
- Go 1.22+
- `psql` client (used by Makefile targets)
- Windows: WSL2 + Docker Desktop (WSL integration)

## Environment Split (by service)
Для локалки можно использовать один `.env`, но для безопасного запуска лучше разносить переменные по сервисам:
- [env/README.md](env/README.md)
- `env/common.env.example`
- `env/api.env.example`
- `env/agent.env.example`
- `env/reporter.env.example`
- `env/bot.env.example`
- `env/db.env.example`

`Makefile` поддерживает такую схему: сначала читает `.env` (backward compatibility), потом `env/common.env`, затем `env/<service>.env`.

## Quick Start (Local)
```bash
make up
<<<<<<< ours
make dbcheck
make migrate
make tables
=======
make migrate
>>>>>>> theirs
make test
make selftest
make logs
```

<<<<<<< ours
Default DSN for Makefile targets:
`postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable`

## Frontend (React + TS + RTK + FSD)
The repository now includes a dedicated frontend app in:
- `frontend/`

Run locally:
```bash
cd frontend
npm.cmd install
npm.cmd run dev
```

Checks:
```bash
npm.cmd run lint
npm.cmd run build
```

Static HTML concept preview:
- `docs/frontend-preview.html`

## Infrastructure Plan (Server + GitHub Pages)
- Repo: `https://github.com/goritskimihail/mudro`
- OS: Ubuntu 22.04 Server (no Desktop).
- Static site/docs: GitHub Pages (default `*.github.io` subdomain until paid domain).
- Runtime on VPS: Docker Compose (API + bot + Postgres + Redis + reverse proxy).
- Email: external SMTP provider (no self-hosted mail).
- Media storage: MinIO (S3-compatible), optional at MVP.
- Access: VS Code Remote-SSH for development and ops.
- Admin panel: optional later (Cockpit or ISPmanager).
- Актуальный чеклист переноса на новый VPS (Ubuntu 24.04):
  [docs/server-transfer-ubuntu24.md](docs/server-transfer-ubuntu24.md)

### Step-by-step Plan
1. GitHub Pages (Docs)
   - Source: `main` + `/root` (repository root).
   - Content: documentation pages (install/config/architecture/FAQ).
   - Result: `https://<owner>.github.io/<repo>` as free domain.

2. VPS Base Setup (Ubuntu 22.04 Server)
   - Create `admin` user, disable root SSH login.
   - SSH keys only, disable password auth.
   - UFW allow: 22, 80, 443.
   - Install Docker + Compose.

3. Project Runtime (Docker Compose)
   - Services: `api`, `bot`, `postgres`, `redis`, `nginx` (or Traefik).
   - `.env` on server: tokens, DB DSN, SMTP, API base URL.
   - Volumes: `postgres_data`, `media_data`.

4. API + Bot
   - Build images from repo or use `go run` in container.
   - Health: `GET /healthz`.
   - Telegram bot: set `TELEGRAM_BOT_TOKEN`, `TELEGRAM_ALLOWED_USERNAME`.

5. Database
   - Postgres with port exposed only to internal network.
   - Run migrations on deploy.
   - Backups: nightly dump to volume or object storage.

6. Media Storage (optional MVP)
   - Add MinIO service, S3-compatible.
   - Bucket for media, credentials in `.env`.
   - Expose via reverse proxy or internal only.

7. Email (registration)
   - Use external SMTP provider.
   - Store SMTP creds in `.env`.
   - Add email verification flow in API.

8. CI/CD (optional)
   - GitHub Actions: build, test, push image.
   - Deploy: `docker compose pull && docker compose up -d`.
   - Secrets stored in GitHub Actions and on server.

9. Monitoring (optional)
   - Add `node_exporter` + basic alerts.
   - Log rotation and disk usage alerts.

## Microservices Roadmap
Концепция декомпозиции, Kafka-потоков и лимитеров описана в:
- [docs/microservices-architecture.md](docs/microservices-architecture.md)

Коротко:
- выделяем `feed-api`, `agent-planner`, `agent-worker`, `reporter`, `telegram-bot`, import-сервисы;
- Kafka как event backbone (`posts/comments/tasks/notifications`);
- RateLimiter на входящий API-трафик и внешние интеграции.
=======
## Importers

### VK -> DB
```bash
go run ./cmd/vkimport -dir ~/vk-export -dsn "postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable"
```

### Telegram export -> JSON (+ optional DB sync)
```bash
go run ./cmd/tgimport -in result.json -out feed_items.json
go run ./cmd/tgimport -in result.json -out feed_items.json -dsn "postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable"
```


>>>>>>> theirs

## Health Loop
```bash
make health
```

Для автономного прогона локального работяги (авто-лог + ретраи + запись в `.codex/state.md`):
```bash
make worker-loop
```
Подробный регламент: [docs/worker-autonomy.md](docs/worker-autonomy.md)

## Agent Queue (MVP)
Простой каркас автономного агента:
- planner читает `.codex/todo.md` и ставит задачи в `agent_queue`
- worker забирает задачи и обрабатывает safe-only типы

```bash
make migrate-agent
make migrate-agent-events
make agent-plan-once
make agent-work
```

Частые safe task kinds в worker:
- `health_check` -> `make test`
- `db_check` -> `make dbcheck`
- `tables_check` -> `make tables`
- `count_posts` -> `make count-posts`

## Repo Layout
- `cmd/api` — HTTP API server (JSON endpoints)
- `cmd/vkimport` — VK raw export -> Postgres (`posts` + `media`)
- `cmd/tgimport` — Telegram export JSON -> feed JSON
- `cmd/bot` — Telegram bot (логи, действия, здоровье за день)
- `migrations/` — SQL migrations (initial schema in `001_init.sql`)
- `internal/` — core packages

## Importers
VK import (default reads `~/vk-export/vk_wall_*.json`):
```bash
go run ./cmd/vkimport \
  --dir ~/vk-export \
  --dsn "postgres://postgres:postgres@localhost:5433/gallery"
```

Telegram export -> JSON feed:
```bash
go run ./cmd/tgimport \
  -in result.json \
  -out feed_items.json \
  -collected-at 2026-02-23T10:00:00Z
```

## Telegram Bot
Set token and run:
```bash
export TELEGRAM_BOT_TOKEN="...your token..."
export TELEGRAM_ALLOWED_USERNAME="sirilarion"
export OPENAI_API_KEY="...your key..."
export API_BASE_URL="http://127.0.0.1:8080"
# optional
# export OPENAI_MODEL="gpt-4.1-mini"
make bot-run
```

Текущие команды бота:
- `/start`
- `/help`
- `/mudro <вопрос>` (ответ LLM с учетом контекста проекта)
- `/now` (краткий итог: достижения/изменения/статус с запуска)
- `/todo` (будущие цели и улучшения)
- `/todoadd <текст>` (добавить цель/улучшение в TODO)
- `/top10` (топ-10 наиболее значимых изменений проекта)
- `/repo` (структура репозитория)
- `/find` (поиск улучшений по репозиторию + авто-добавление важного в TODO)
- `/time` (суммарное время работы + время генерации ответов)
- `/rab` (автовыполнение простых задач из TODO, перенос в DONE, допоиск улучшений)
- `/memento` (полная синхронизация памяти проекта и снимок структуры)
- `/tglog` (история управления, что запускалось из Telegram)
- `/chat on|off|status` (режим обычного чата без команд)
- `/reportnow` (мгновенно отправить отчет через reporter-бота)
- `/feed5` (вывод 5 постов из `GET /api/front`)

## Reporter Bot (separate)
```bash
export REPORT_BOT_TOKEN="...second bot token..."
# optional if fallback should not be used:
# export REPORT_CHAT_ID="123456789"
# optional:
# export REPORT_INTERVAL_MIN="30"
make report-run
```
- `/health` (состояние сейчас + поломки/успехи за день + git-итоги за день)
- `/logs`
- `/actions10`
- `/actions1h`
- `/commits3`

## Configuration
Override DSN for Make targets:
```bash
DSN="postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable" make dbcheck
```

## HTTP API
Run the API server:
```bash
API_ADDR=":8080" DSN="postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable" go run ./cmd/api
```

Rate limiter (входящие запросы):
- `API_RATE_LIMIT_RPS` (default `20`)
- `API_RATE_LIMIT_BURST` (default `40`)
- `REDIS_RATE_LIMIT_ENABLED=true` переключает лимитер на Redis backend (`REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`)
- если `API_RATE_LIMIT_RPS=0`, лимитер выключен

Kafka events (agent runtime):
- при `KAFKA_ENABLED=true` `cmd/agent` публикует lifecycle-события задач в `KAFKA_TOPIC_TASKS`
- брокеры задаются через `KAFKA_BROKERS` (comma-separated)

Endpoints:
- `GET /healthz` → `{ "status": "ok" }`
- `GET /feed?limit=20` (простой HTML-шаблон ленты с кнопкой "Загрузить еще")
- `GET /api/front?limit=50` (одним запросом: мета + лента для фронта)
- `GET /api/posts?limit=50`
- `GET /api/posts?page=2` (page-based, 1-indexed)
- Cursor pagination: `GET /api/posts?before_ts=2026-02-24T12:00:00Z&before_id=123`

Response wrapper (page-based):
```json
{
  "page": 2,
  "limit": 50,
  "items": [ ... ]
}
```
## Notes
- Для этого проекта лента VK рассматривается как архив: повторные обновления VK не требуются.
- Подготовлена схема персональных лайков: `accounts` + `post_account_likes` (one-like-per-account-per-post).

## Operations
- Mission and guardrails: `Mission.md`
- OpenClaw integration notes: `docs/openclaw-integration.md`
- Draft process policy: `BIBLE.proposed.md`
