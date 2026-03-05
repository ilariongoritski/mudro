# mudro

Go backend (Postgres) for unified VK + Telegram feed.

## Requirements
- Docker + Docker Compose
- Go 1.22+ (recommended)
- Windows: WSL2 + Docker Desktop (WSL integration)

## Quick start
```bash
make up
make migrate
make test
make selftest
make logs
```

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






## Notes
- Для этого проекта лента VK рассматривается как архив: повторные обновления VK не требуются.
- Подготовлена схема персональных лайков: `accounts` + `post_account_likes` (one-like-per-account-per-post).

## Operations
- Mission and guardrails: `Mission.md`
- OpenClaw integration notes: `docs/openclaw-integration.md`
- Draft process policy: `BIBLE.proposed.md`
