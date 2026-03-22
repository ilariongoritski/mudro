# Локальный работяга: автономный режим (почти без ручного ведения)

Этот документ — практический гайд, чтобы новый локальный работяга мог быстро войти в проект и работать по понятному циклу с минимальным ручным управлением.

## 1) Что нужно знать за 2 минуты
- Репозиторий: `mudro` (Go + Postgres + Docker Compose).
- Главный цикл проверки: health loop (`make up -> ps -> dbcheck -> migrate -> tables -> test -> count-posts`).
- Канонический DSN: `postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable`.
- Память проекта: `.codex/todo.md`, `.codex/done.md`, `.codex/top10.md`, `.codex/memory.json`, `.codex/time_runtime.json`, `.codex/tg_control.jsonl`.

## 2) Быстрый старт (обязательный минимум)
1. Открыть корень проекта: `cd ~/projects/mudro` (или актуальный локальный путь).
2. Прочитать память:
   - `.codex/todo.md`
   - `.codex/done.md`
   - `.codex/top10.md`
   - `.codex/memory.json`
   - `.codex/time_runtime.json`
   - `.codex/tg_control.jsonl`
3. Запустить автономный цикл:
   - `make worker-loop`

Скрипт сам:
- создает лог `.codex/logs/YYYYMMDD-HHMM/index.md`;
- выполняет health loop с максимум 2 попытками на шаг;
- дописывает краткий статус в `.codex/state.md`.

## 3) Что делать после health loop
Если health loop успешен:
1. Выбрать 1-2 верхние задачи из `.codex/todo.md` (P0/P1 в приоритете).
2. Для каждой задачи:
   - сделать минимальный безопасный diff;
   - прогнать релевантные проверки (`make test`, точечные тесты, либо профильные команды);
   - обновить `.codex/done.md` и при необходимости `.codex/top10.md`;
   - обновить `.codex/state.md`.
3. Сделать `git commit` с понятным сообщением.

Если health loop неуспешен:
1. Остановиться после 2 попыток упавшего шага.
2. Подготовить вопрос человеку:
   - точная команда;
   - точная ошибка;
   - путь к логу `.codex/logs/<run>/index.md`.

## 4) Безопасные границы (чтобы не навредить)
Нельзя без явного подтверждения:
- `drop/truncate/reset` данных;
- `docker compose down -v`;
- массовые удаления;
- изменение публичного JSON API-контракта;
- изменение DDL миграций (кроме comment-only фиксов мусорных комментариев).

Разрешено без подтверждения:
- повтор упавшей команды 1 раз;
- мягкий рестарт (`make down`, `make up`, `docker compose ps`) без удаления volume;
- документальные и логирующие изменения в `.codex/*`.

## 5) Карта проекта (для автономного ориентирования)
- `cmd/api` — HTTP API и рендер `/feed`.
- `cmd/bot` — Telegram бот (управление/память/отчеты).
- `cmd/reporter` — отдельный репортер-бот с digest.
- `cmd/agent` + `internal/agent` — planner/worker и очередь задач.
- `cmd/tgimport`, `cmd/tgload`, `cmd/vkimport`, `cmd/tgcommentsimport` — импорт контента.
- `internal/api`, `internal/bot`, `internal/reporter`, `internal/config` — доменная логика.
- `migrations/` — SQL миграции.
- `env/*.env.example` — шаблоны конфигурации по сервисам.

## 6) Режим «почти автономно при лимитах»
Чтобы работать стабильно при ограниченном времени/токенах:
1. Делать короткие прогоны и маленькие PR (1 смысловая цель на коммит).
2. Всегда начинать с `make worker-loop` (или руками health loop, если нужен контроль).
3. Любую найденную проблему сначала фиксировать в `.codex/todo.md`, потом чинить.
4. После фикса — сразу переносить пункт в `.codex/done.md` с измеримым эффектом.
5. Если изменение архитектурно значимо — обновлять `.codex/top10.md`.

## 7) Готовые команды
- Полный health loop (авто-лог + ретраи):
  - `make worker-loop`
- Локальный подъем:
  - `make up && docker compose ps`
- БД/миграции:
  - `make dbcheck && make migrate && make tables`
- Тесты:
  - `make test`
- Санити:
  - `make count-posts`

## 8) Критерий «работяга живой»
- БД-контейнер healthy;
- `select 1` проходит;
- миграции применяются без ошибок;
- таблица `posts` существует;
- тесты проходят;
- `count(posts)` выполняется.

Если `count = 0` — это допустимо (данные могут быть еще не импортированы).

## 9) Orchestration: Claude Opus + Codex
Режим для сложных задач:
- `Claude Opus` готовит plan/review/draft (не пишет напрямую в tracked файлы репозитория).
- `Codex` применяет изменения локально, запускает проверки и ведет финальный diff.

Языковая политика:
- internal prompts/handoffs между агентами — English;
- пользовательские объяснения и отчеты — русский.

Логи и память:
- `.codex/logs/<run>/index.md`
- `.codex/state.md`
- `.codex/todo.md`
- `.codex/done.md`
- при необходимости: `.codex/time_runtime.json`, `.codex/tg_control.jsonl`

Новый отдельный лог-формат не вводится.

## 10) Bootstrap fresh run log
- Use `make orchestration-log-init RUN_ID=<run> TASK="<short task>"` to seed `.codex/logs/<run>/index.md`.
- The generated log keeps the English sections required by the orchestration contract.

## 12) UTF-8 and local-only files
- Use explicit UTF-8 for text I/O in PowerShell and shell commands. If output looks garbled, rerun with explicit encoding or use WSL/bash.
- Keep auxiliary local files, downloads, caches, and tool installs under `D:\mudr\_mudro-local` instead of the tracked repo tree.
