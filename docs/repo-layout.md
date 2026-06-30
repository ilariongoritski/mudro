# Repository Layout

Этот документ фиксирует минимально вменяемую структуру репозитория без массовых переносов и без ломки текущего кода.

## Канонические корневые папки

- `api/` — serverless/adaptor entrypoints
- `cmd/` — только канонические top-level CLI, сейчас `cmd/mudro`
- `contracts/` — контрактные схемы и форматы
- `data/` — локальные сырые выгрузки и приватные входные данные, не коммитятся
- `docs/` — постоянная документация проекта
- `e2e/` — smoke/e2e тесты
- `env/` — примеры env-профилей и пояснения
- `frontend/` — React/TypeScript UI
- `internal/` — основной Go-код проекта
- `migrations/` — SQL-миграции
- `out/` — legacy generated output, не расширять без необходимости
- `output/` — все новые локальные generated-артефакты и временные результаты
- `ops/` — runbooks, compose-профили, ops-скрипты и env-шаблоны для эксплуатации
- `pkg/` — экспортируемые/вспомогательные пакеты
- `scripts/` — операционные и dev-скрипты
- `services/` — long-running runtime сервисы (`feed-api`, `agent`, `bot`)
- `tools/` — одноразовые и операционные CLI (import/backfill/maintenance)
- `tmp/` — временные локальные файлы, не коммитятся

## Runtime и CLI baseline

- Активный runtime-контур живет в `services/*`.
- Одноразовые и операционные команды живут в `tools/*`.
- Root `cmd/mudro` оставлен как канонический агрегирующий CLI.
- Legacy forwarding entrypoints и `legacy/old/*` удалены; не добавлять на них ссылки в Makefile/CI/runbook.

## Что считать "грязью"

Нежелательно класть в корень репозитория:

- временные `.go` helper-файлы
- разовые `.sh`/`.ps1` пробы
- screenshots, `.log`, `.json`, `.csv`, `.pdf`, `.docx`
- локальные node/python toolchains

Для этого использовать:

- `output/logs/`
- `output/playwright/`
- `output/db/`
- `output/import/`
- `output/doc/`
- `output/deploy/`
- `output/ssh-tool/`
- `tmp/`

## Простое правило размещения

1. Это код проекта или постоянная документация?
   - Класть в `cmd/`, `internal/`, `frontend/`, `docs/`, `scripts/`
2. Это входные исходные данные импорта?
   - Класть в `data/`
3. Это одноразовый результат, лог, скриншот, сверка, экспорт, probe?
   - Класть в `output/`
4. Это совсем временный черновик на текущую сессию?
   - Класть в `tmp/`
