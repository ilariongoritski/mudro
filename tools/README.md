# Tools

Каталог `tools/` содержит локальные и операционные инструменты проекта MUDRO. Это не long-running runtime-сервисы, а вспомогательные контуры для импорта данных, backfill, обслуживания и локальной агентной автоматизации.

## Основные группы

- `importers/`
  Импорт данных из Telegram, VK и связанных экспортов.
- `backfill/`
  Дозаполнение уже загруженных данных, например комментариев и медиа.
- `maintenance/`
  Сервисные утилиты для merge, dedupe и исторического обслуживания данных.
- `opus-gateway/`
  Локальный HTTP sidecar для запуска задач через Anthropic `Opus` на вашем `ANTHROPIC_API_KEY`, scoped к этому репозиторию.

## Opus Gateway

`tools/opus-gateway/` нужен, когда вы хотите запускать отдельный локальный агентный контур рядом с MUDRO, не смешивая его со встроенными субагентами текущего Codex-чата.

Что он делает:
- поднимает локальный HTTP сервис на `127.0.0.1:8788`
- принимает `GET /healthz` и `POST /v1/run`
- ограничивает рабочую директорию корнем репозитория
- поддерживает `read-only` и `edit` режимы
- может включать строго allowlisted `Bash`

Базовый запуск из корня репозитория:

```bash
npm run opus-gateway:install
npm run opus-gateway
```

См. также:
- [`tools/opus-gateway/README.md`](./opus-gateway/README.md)
- [`README.md`](../README.md)
- [`docs/agent-workflows.md`](../docs/agent-workflows.md)
