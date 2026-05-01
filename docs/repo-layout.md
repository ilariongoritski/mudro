# Repository Layout

Этот документ фиксирует минимально вменяемую структуру репозитория без массовых переносов и без ломки текущего кода.

## Канонические корневые папки

- `api/` — serverless/adaptor entrypoints
- `cmd/` — исполняемые команды и one-shot CLI
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

## Миграционный слой совместимости

- Legacy runtime entrypoints are kept in `legacy/old/cmd-runtime/*`.
- Активный runtime-контур живет только в `services/*`.
- `cmd/*` используется как временный forwarding-слой для CLI в `tools/*`.
- Legacy `reporter` вынесен в `legacy/old/services/reporter-old` и не участвует в default runtime.

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
