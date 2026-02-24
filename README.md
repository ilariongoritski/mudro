# MUDRO — Go backend for a unified VK + Telegram feed

MUDRO is a Go + Postgres backend with importers for VK export data and Telegram export JSON. It also includes a Telegram bot and a health-check command runner.

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
- `cmd/vkimport` — VK raw export -> Postgres (`posts` + `media`)
- `cmd/tgimport` — Telegram export JSON -> feed JSON
- `cmd/bot` — Telegram bot (commands handler)
- `cmd/server/kserver` — Telegram bot with `/health` command
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
go run ./cmd/bot
```

Health bot (runs local health loop on `/health`):
```bash
export TELEGRAM_BOT_TOKEN="...your token..."
go run ./cmd/server/kserver
```

## Configuration
Override DSN for Make targets:
```bash
DSN="postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable" make dbcheck
```




