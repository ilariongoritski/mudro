# Microservice Layout (Target)

Этот документ фиксирует целевую структуру `mudro` как единого микросервисного проекта с безопасной миграцией от монолитного дерева.

## Принцип разделения

1. Runtime-сервисы живут в `services/*` и запускаются независимо.
2. Общая доменная логика и контракты остаются в `internal/*` и `contracts/*`.
3. Одноразовые утилиты импорта/обслуживания живут в `tools/*` (переходно — часть в `cmd/*`).
4. `cmd/*` используется только как временный forwarding-слой для CLI в `tools/*`.

## Целевая карта директорий

```text
services/
  feed-api/
    cmd/
    app/
  agent/
    cmd/
    app/
  bot/
    cmd/
    app/
tools/
  importers/           # tg/vk import CLI
  backfill/            # comment/media backfill
  maintenance/         # dedupe, merge, memento, checks

internal/
  api/
  agent/
  bot/
  config/
  ...

contracts/
  http/
  events/
```

## Границы сервисов

- `feed-api`: HTTP read/query facade + health + public contract.
- `agent`: planner/worker orchestration и lifecycle событий задач.
- `bot`: Telegram control-plane для оператора.
- `reporter-old`: вынесен в `legacy/old/services/reporter-old`, не является частью default runtime.

## Правила расширения

1. Новый runtime-сервис добавляется только в `services/<name>`.
2. Новый CLI-инструмент добавляется в `tools/*`, а не в корень.
3. Общие функции не дублируются между сервисами: вынос в `internal/*`.
4. Публичные форматы фиксируются в `contracts/*` до интеграции между сервисами.

## Переходный статус (сейчас)

- Активные runtime entrypoints: `services/feed-api`, `services/agent`, `services/bot`.
- `cmd/api|agent|bot|reporter` перенесены в `legacy/old/cmd-runtime/*`.
- Импортеры/backfill/maintenance перенесены в `tools/*`, а `cmd/*` оставлен как временный forwarding-слой.
