# ADR в MUDRO

## Когда создавать ADR
ADR нужен только для решений, которые меняют устойчивые правила проекта:

- модель данных
- импортные пайплайны
- публичный API-контракт
- инфраструктуру и runtime
- security/hardening политику
- правила хранения и канонизации контента

Если изменение локальное и не переживет одну задачу, отдельный ADR не нужен.

## Именование
Формат имени файла:

`docs/adr/0001-short-title.md`

Нумерация монотонная. Заголовок короткий и технический.

## Минимальный шаблон
```md
# ADR 0001: Short Title

- Status: proposed
- Date: YYYY-MM-DD

## Context
Какой был исходный контекст, ограничение или проблема.

## Decision
Что именно решили сделать.

## Alternatives
- Вариант A
- Вариант B

## Consequences
- Что упрощается
- Какие компромиссы появляются
- Что нужно обновить в коде, docs или runbook
```

## Для MUDRO особенно уместны ADR по темам
- `media_assets`, `post_media_links`, `comment_media_links`
- comment model и parent/reactions graph
- `VK snapshot-only`
- VPS self-hosted frontend через `nginx`
- hardening Postgres и сервисных ролей
