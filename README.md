# MUDRO — Go backend for a unified VK + Telegram feed

MUDRO is a Go + Postgres backend with importers for VK export data and Telegram export JSON. It also includes a Telegram bot for operational summaries.

## Requirements
- Docker + Docker Compose
- Go 1.22+
- `psql` client (used by Makefile targets)
- Windows: WSL2 + Docker Desktop (WSL integration)

## Quick Start (Local)
```bash
make up
make dbcheck
make migrate
make tables
make test
make logs
```

Default DSN for Makefile targets:
`postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable`

## Health Loop
```bash
make health
```

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
- `/time` (суммарное время работы: за день и всего)
- `/rab` (автовыполнение простых задач из TODO, перенос в DONE, допоиск улучшений)
- `/memento` (полная синхронизация памяти проекта и снимок структуры)
- `/feed5` (вывод 5 постов из `GET /api/front`)
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

Endpoints:
- `GET /healthz` → `{ "status": "ok" }`
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
