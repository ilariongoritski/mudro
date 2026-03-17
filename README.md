# MUDRO вЂ” Go backend for a unified VK + Telegram feed

MUDRO is a Go + Postgres backend with importers for VK export data and Telegram export JSON. It also includes a Telegram bot for operational summaries.

## Requirements
- Docker + Docker Compose
- Go 1.22+
- `psql` client (used by Makefile targets)
- Windows: WSL2 + Docker Desktop (WSL integration)

## Repository Layout
- Runtime code: `cmd/`, `internal/`, `frontend/`, `api/`
- SQL and config examples: `migrations/`, `env/`
- Stable docs: `docs/`
- Local raw imports: `data/`
- Local generated artifacts only: `output/`, `tmp/`, `out/`

Короткое правило: если файл не является исходным кодом, миграцией, env-примером или постоянной документацией, он должен жить в `output/` или `tmp/`, а не в корне репозитория. Подробности: [docs/repo-layout.md](docs/repo-layout.md)

## Environment Split (by service)
Р”Р»СЏ Р»РѕРєР°Р»РєРё РјРѕР¶РЅРѕ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊ РѕРґРёРЅ `.env`, РЅРѕ РґР»СЏ Р±РµР·РѕРїР°СЃРЅРѕРіРѕ Р·Р°РїСѓСЃРєР° Р»СѓС‡С€Рµ СЂР°Р·РЅРѕСЃРёС‚СЊ РїРµСЂРµРјРµРЅРЅС‹Рµ РїРѕ СЃРµСЂРІРёСЃР°Рј:
- [env/README.md](env/README.md)
- `env/common.env.example`
- `env/api.env.example`
- `env/agent.env.example`
- `env/reporter.env.example`
- `env/bot.env.example`
- `env/db.env.example`

`Makefile` РїРѕРґРґРµСЂР¶РёРІР°РµС‚ С‚Р°РєСѓСЋ СЃС…РµРјСѓ: СЃРЅР°С‡Р°Р»Р° С‡РёС‚Р°РµС‚ `.env` (backward compatibility), РїРѕС‚РѕРј `env/common.env`, Р·Р°С‚РµРј `env/<service>.env`.

Minimal runtime config rules:
- `MUDRO_ENV=development` for local runs; `production` or `staging` for server/runtime services.
- `REPORT_CHAT_ID` is mandatory for `make report-run` and `cmd/reporter`.
- Superuser DSN (`user=postgres`) is allowed only for local dev hosts (`localhost`, `127.0.0.1`, `db`); production/staging must use a separate app role.

## Quick Start (Local)
```bash
make up
make dbcheck
make migrate
make tables
make migrate
make test
make selftest
make logs
```

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

Server-side static rollout on VPS:
- build locally or on the server: `cd frontend && npm.cmd run build`
- deploy static site + reverse proxy on VPS:
  - `bash scripts/ops/deploy_vps_frontend.sh`

Result:
- frontend is served by `nginx` on `:80`
- `/api`, `/media`, `/healthz` are proxied to local `mudro-api` on `127.0.0.1:8080`

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
- РђРєС‚СѓР°Р»СЊРЅС‹Р№ С‡РµРєР»РёСЃС‚ РїРµСЂРµРЅРѕСЃР° РЅР° РЅРѕРІС‹Р№ VPS (Ubuntu 24.04):
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

Security note for VPS:
- Postgres must not stay public on `0.0.0.0:5433` with default credentials.
- On server bind DB to loopback only and run application services under a dedicated non-superuser role (for example `mudro_app`).

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
РљРѕРЅС†РµРїС†РёСЏ РґРµРєРѕРјРїРѕР·РёС†РёРё, Kafka-РїРѕС‚РѕРєРѕРІ Рё Р»РёРјРёС‚РµСЂРѕРІ РѕРїРёСЃР°РЅР° РІ:
- [docs/microservices-architecture.md](docs/microservices-architecture.md)

РљРѕСЂРѕС‚РєРѕ:
- РІС‹РґРµР»СЏРµРј `feed-api`, `agent-planner`, `agent-worker`, `reporter`, `telegram-bot`, import-СЃРµСЂРІРёСЃС‹;
- Kafka РєР°Рє event backbone (`posts/comments/tasks/notifications`);
- RateLimiter РЅР° РІС…РѕРґСЏС‰РёР№ API-С‚СЂР°С„РёРє Рё РІРЅРµС€РЅРёРµ РёРЅС‚РµРіСЂР°С†РёРё.
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



## Health Loop
```bash
make health
```

Р”Р»СЏ Р°РІС‚РѕРЅРѕРјРЅРѕРіРѕ РїСЂРѕРіРѕРЅР° Р»РѕРєР°Р»СЊРЅРѕРіРѕ СЂР°Р±РѕС‚СЏРіРё (Р°РІС‚Рѕ-Р»РѕРі + СЂРµС‚СЂР°Рё + Р·Р°РїРёСЃСЊ РІ `.codex/state.md`):
```bash
make worker-loop
```
РџРѕРґСЂРѕР±РЅС‹Р№ СЂРµРіР»Р°РјРµРЅС‚: [docs/worker-autonomy.md](docs/worker-autonomy.md)

## Spec-Driven Workflow
Repo-native adaptation of the Skaro article for `mudro`:
- [docs/skaro-mudro-adaptation.md](docs/skaro-mudro-adaptation.md)
- [docs/adr/README.md](docs/adr/README.md)
- [.codex/templates/task-spec.md](.codex/templates/task-spec.md)

## Agent Queue (MVP)
РџСЂРѕСЃС‚РѕР№ РєР°СЂРєР°СЃ Р°РІС‚РѕРЅРѕРјРЅРѕРіРѕ Р°РіРµРЅС‚Р°:
- planner С‡РёС‚Р°РµС‚ `.codex/todo.md` Рё СЃС‚Р°РІРёС‚ Р·Р°РґР°С‡Рё РІ `agent_queue`
- worker Р·Р°Р±РёСЂР°РµС‚ Р·Р°РґР°С‡Рё Рё РѕР±СЂР°Р±Р°С‚С‹РІР°РµС‚ safe-only С‚РёРїС‹

```bash
make migrate-agent
make migrate-agent-events
make agent-plan-once
make agent-work
```

Р§Р°СЃС‚С‹Рµ safe task kinds РІ worker:
- `health_check` -> `make test`
- `db_check` -> `make dbcheck`
- `tables_check` -> `make tables`
- `count_posts` -> `make count-posts`

## Repo Layout
- `cmd/api` вЂ” HTTP API server (JSON endpoints)
- `cmd/vkimport` вЂ” VK raw export -> Postgres (`posts` + `media`)
- `cmd/tgimport` вЂ” Telegram export JSON -> feed JSON
- `cmd/bot` вЂ” Telegram bot (Р»РѕРіРё, РґРµР№СЃС‚РІРёСЏ, Р·РґРѕСЂРѕРІСЊРµ Р·Р° РґРµРЅСЊ)
- `migrations/` вЂ” SQL migrations (initial schema in `001_init.sql`)
- `internal/` вЂ” core packages

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

РўРµРєСѓС‰РёРµ РєРѕРјР°РЅРґС‹ Р±РѕС‚Р°:
- `/start`
- `/help`
- `/mudro <РІРѕРїСЂРѕСЃ>` (РѕС‚РІРµС‚ LLM СЃ СѓС‡РµС‚РѕРј РєРѕРЅС‚РµРєСЃС‚Р° РїСЂРѕРµРєС‚Р°)
- `/now` (РєСЂР°С‚РєРёР№ РёС‚РѕРі: РґРѕСЃС‚РёР¶РµРЅРёСЏ/РёР·РјРµРЅРµРЅРёСЏ/СЃС‚Р°С‚СѓСЃ СЃ Р·Р°РїСѓСЃРєР°)
- `/todo` (Р±СѓРґСѓС‰РёРµ С†РµР»Рё Рё СѓР»СѓС‡С€РµРЅРёСЏ)
- `/todoadd <С‚РµРєСЃС‚>` (РґРѕР±Р°РІРёС‚СЊ С†РµР»СЊ/СѓР»СѓС‡С€РµРЅРёРµ РІ TODO)
- `/top10` (С‚РѕРї-10 РЅР°РёР±РѕР»РµРµ Р·РЅР°С‡РёРјС‹С… РёР·РјРµРЅРµРЅРёР№ РїСЂРѕРµРєС‚Р°)
- `/repo` (СЃС‚СЂСѓРєС‚СѓСЂР° СЂРµРїРѕР·РёС‚РѕСЂРёСЏ)
- `/find` (РїРѕРёСЃРє СѓР»СѓС‡С€РµРЅРёР№ РїРѕ СЂРµРїРѕР·РёС‚РѕСЂРёСЋ + Р°РІС‚Рѕ-РґРѕР±Р°РІР»РµРЅРёРµ РІР°Р¶РЅРѕРіРѕ РІ TODO)
- `/time` (СЃСѓРјРјР°СЂРЅРѕРµ РІСЂРµРјСЏ СЂР°Р±РѕС‚С‹ + РІСЂРµРјСЏ РіРµРЅРµСЂР°С†РёРё РѕС‚РІРµС‚РѕРІ)
- `/rab` (Р°РІС‚РѕРІС‹РїРѕР»РЅРµРЅРёРµ РїСЂРѕСЃС‚С‹С… Р·Р°РґР°С‡ РёР· TODO, РїРµСЂРµРЅРѕСЃ РІ DONE, РґРѕРїРѕРёСЃРє СѓР»СѓС‡С€РµРЅРёР№)
- `/memento` (РїРѕР»РЅР°СЏ СЃРёРЅС…СЂРѕРЅРёР·Р°С†РёСЏ РїР°РјСЏС‚Рё РїСЂРѕРµРєС‚Р° Рё СЃРЅРёРјРѕРє СЃС‚СЂСѓРєС‚СѓСЂС‹)
- `/tglog` (РёСЃС‚РѕСЂРёСЏ СѓРїСЂР°РІР»РµРЅРёСЏ, С‡С‚Рѕ Р·Р°РїСѓСЃРєР°Р»РѕСЃСЊ РёР· Telegram)
- `/chat on|off|status` (СЂРµР¶РёРј РѕР±С‹С‡РЅРѕРіРѕ С‡Р°С‚Р° Р±РµР· РєРѕРјР°РЅРґ)
- `/reportnow` (РјРіРЅРѕРІРµРЅРЅРѕ РѕС‚РїСЂР°РІРёС‚СЊ РѕС‚С‡РµС‚ С‡РµСЂРµР· reporter-Р±РѕС‚Р°)
- `/feed5` (РІС‹РІРѕРґ 5 РїРѕСЃС‚РѕРІ РёР· `GET /api/front`)

## Reporter Bot (separate)
```bash
export REPORT_BOT_TOKEN="...second bot token..."
export REPORT_CHAT_ID="123456789"
# optional:
# export REPORT_INTERVAL_MIN="30"
make report-run
```
- `/health` (СЃРѕСЃС‚РѕСЏРЅРёРµ СЃРµР№С‡Р°СЃ + РїРѕР»РѕРјРєРё/СѓСЃРїРµС…Рё Р·Р° РґРµРЅСЊ + git-РёС‚РѕРіРё Р·Р° РґРµРЅСЊ)
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
MUDRO_ENV="development" API_ADDR=":8080" DSN="postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable" go run ./cmd/api
```

Rate limiter (РІС…РѕРґСЏС‰РёРµ Р·Р°РїСЂРѕСЃС‹):
- `API_RATE_LIMIT_RPS` (default `20`)
- `API_RATE_LIMIT_BURST` (default `40`)
- `REDIS_RATE_LIMIT_ENABLED=true` РїРµСЂРµРєР»СЋС‡Р°РµС‚ Р»РёРјРёС‚РµСЂ РЅР° Redis backend (`REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`)
- РµСЃР»Рё `API_RATE_LIMIT_RPS=0`, Р»РёРјРёС‚РµСЂ РІС‹РєР»СЋС‡РµРЅ

Kafka events (agent runtime):
- РїСЂРё `KAFKA_ENABLED=true` `cmd/agent` РїСѓР±Р»РёРєСѓРµС‚ lifecycle-СЃРѕР±С‹С‚РёСЏ Р·Р°РґР°С‡ РІ `KAFKA_TOPIC_TASKS`
- Р±СЂРѕРєРµСЂС‹ Р·Р°РґР°СЋС‚СЃСЏ С‡РµСЂРµР· `KAFKA_BROKERS` (comma-separated)

Endpoints:
- `GET /healthz` в†’ `{ "status": "ok" }`
- `GET /feed?limit=20` (РїСЂРѕСЃС‚РѕР№ HTML-С€Р°Р±Р»РѕРЅ Р»РµРЅС‚С‹ СЃ РєРЅРѕРїРєРѕР№ "Р—Р°РіСЂСѓР·РёС‚СЊ РµС‰Рµ")
- `GET /api/front?limit=50` (РѕРґРЅРёРј Р·Р°РїСЂРѕСЃРѕРј: РјРµС‚Р° + Р»РµРЅС‚Р° РґР»СЏ С„СЂРѕРЅС‚Р°)
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
- Р”Р»СЏ СЌС‚РѕРіРѕ РїСЂРѕРµРєС‚Р° Р»РµРЅС‚Р° VK СЂР°СЃСЃРјР°С‚СЂРёРІР°РµС‚СЃСЏ РєР°Рє Р°СЂС…РёРІ: РїРѕРІС‚РѕСЂРЅС‹Рµ РѕР±РЅРѕРІР»РµРЅРёСЏ VK РЅРµ С‚СЂРµР±СѓСЋС‚СЃСЏ.
- РџРѕРґРіРѕС‚РѕРІР»РµРЅР° СЃС…РµРјР° РїРµСЂСЃРѕРЅР°Р»СЊРЅС‹С… Р»Р°Р№РєРѕРІ: `accounts` + `post_account_likes` (one-like-per-account-per-post).

## Operations
- Mission and guardrails: `Mission.md`
- OpenClaw integration notes: `docs/openclaw-integration.md`
- Draft process policy: `BIBLE.proposed.md`
