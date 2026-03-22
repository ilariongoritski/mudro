# Orchestration: Claude Opus + Codex

Цель: зафиксировать рабочий контур, где `Claude Opus` готовит план/ревью/черновики, а `Codex` применяет изменения локально и отвечает за итоговый diff в репозитории.

## Роли
- `Claude Opus`:
  - planning для сложных задач;
  - code review и risk notes;
  - draft diff / implementation outline;
  - acceptance criteria для проверки.
- `Codex`:
  - control plane и source of truth по репозиторию;
  - локальное применение изменений;
  - запуск тестов/health loop;
  - финальная проверка перед merge в `main`.

## Язык коммуникации
- Межагентные handoff и prompts: только English.
- Пользовательские объяснения и отчеты: русский.

### Internal handoff template (English)
```text
Task: <one-line goal>
Scope: <files/subsystems>
Constraints: <safety/performance/contracts>
Expected output:
1) Plan (short)
2) Risks
3) Draft changes
4) Validation checklist
```

## Логи и память (без нового формата)
Используется только существующий `.codex/*` контур:
- `.codex/logs/<run>/index.md` — основной лог прогона;
- `.codex/state.md` — короткий статус и next step;
- `.codex/todo.md` — актуальные задачи;
- `.codex/done.md` — завершенные задачи;
- при необходимости: `.codex/time_runtime.json`, `.codex/tg_control.jsonl`.

Новый отдельный формат логов не вводится.

## Базовый цикл для сложной задачи
1. `Codex` формирует English prompt для `Claude Opus`.
2. `Claude Opus` возвращает plan + draft + risks.
3. `Codex` применяет изменения локально в рабочей ветке/worktree.
4. `Codex` запускает релевантные проверки:
   - backend/db: `go test ./...`, `make dbcheck`, `make tables`, при необходимости `make health`;
   - frontend: `cd frontend && npm.cmd run build`, при UI-изменениях также lint/smoke.
5. `Codex` фиксирует результат в `.codex/logs/...` и `.codex/state.md`.
6. Только после валидации изменения могут идти в `main`.

## Важные ограничения
- `Claude Opus` не пишет напрямую в tracked файлы репозитория.
- Все изменения в коде/конфигах делает локальный `Codex`.
- Деструктивные операции по БД и схеме — только с явным подтверждением человека.

## Run log bootstrap
- Use `make orchestration-log-init RUN_ID=<run> TASK="<short task>"` to seed a new `.codex/logs/<run>/index.md`.
- The helper writes English sections for `Context`, `Task`, `Claude Draft`, `Codex Apply`, `Validation`, and `Handoff`.

## 11) UTF-8 and local-only files
- For shell and file I/O that reads or writes text, use explicit UTF-8. In PowerShell prefer `Get-Content -Encoding UTF8` and `Set-Content -Encoding UTF8`; if output looks garbled, rerun it with explicit encoding or use WSL/bash.
- Keep auxiliary local files, downloads, caches, and tool installs under `D:\mudr\_mudro-local` instead of the tracked repo tree.

## 12) Skaro UI as orchestration panel
- Use `Skaro UI` as the lightweight browser panel for planning, progress tracking, review status, and token usage.
- Keep the real Claude API key only in `D:\mudr\_mudro-local\skaro\claude.env` or process env vars.
- Do not store live credentials in `.skaro/secrets.yaml`.
- The tracked project config may reference the Claude-compatible endpoint, but the launcher is responsible for exporting `ANTHROPIC_API_KEY` and `ANTHROPIC_BASE_URL` before `skaro ui` or `skaro status`.
- If the Claude usage proxy is ever exposed beyond loopback, require `MUDRO_CLAUDE_PROXY_TOKEN` on non-local requests so the proxy does not become an unauthenticated Claude gateway.
