---
name: mudro-import-data
description: Импорт данных (VK/TG посты, комментарии, медиа) в локальную или серверную БД
---

# Skill: Импорт данных Mudro

## Доступные импортеры

| Команда | Что делает | Источник |
|---------|-----------|----------|
| `go run ./cmd/vkimport -dir <path> -dsn <dsn>` | VK wall export → posts | JSON файлы `vk_wall_*.json` |
| `go run ./cmd/tgimport -in result.json -out feed.json -dsn <dsn>` | TG export → posts | `result.json` от Telegram |
| `go run ./cmd/tgcsvimport -in <csv> -dsn <dsn>` | TG CSV посты | CSV с views/comments/reactions |
| `go run ./cmd/tgcommentscsvimport -in <csv> -dsn <dsn>` | TG CSV комментарии | CSV discussion messages |
| `go run ./cmd/tgcommentmediaimport -dir <dir> -dsn <dsn>` | Media комментариев | Директория с файлами |
| `go run ./cmd/mediabackfill` | Backfill normalized media | Из legacy JSONB в `media_assets` |
| `go run ./cmd/commentbackfill` | Backfill комментариев | Из legacy в нормализованную модель |

## Канонический DSN
```
postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable
```

## Порядок импорта (полный цикл)
1. `make migrate-all` — применить все миграции
2. VK import (если нужен): `go run ./cmd/vkimport ...`
3. TG import: `go run ./cmd/tgcsvimport ...` или `go run ./cmd/tgimport ...`
4. TG comments: `go run ./cmd/tgcommentscsvimport ...`
5. Media backfill: `go run ./cmd/mediabackfill`
6. Проверка: `select source, count(*) from posts group by source;`

## Важно
- VK — snapshot-only, повторный импорт не нужен
- TG importers используют upsert, повторный запуск безопасен
- Все команды выполнять из корня репозитория через WSL
