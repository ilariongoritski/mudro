# Интеграция mudro с Клавботом (OpenClaw)

Цель: запускать безопасные циклы `patch -> selftest -> commit -> PR` без прямого доступа к SSH из агента.

## Минимальный контракт команд
Клавботу нужны только ограниченные команды:

1. `repo.snapshot`
   - читает: `README.md`, `Makefile`, `go.mod`, `cmd/*`, `internal/*`, `migrations/*`, `.github/workflows/*`.
2. `repo.selftest`
   - выполняет `make selftest`.
3. `repo.diff`
   - возвращает `git diff --stat` и `git diff`.
4. `repo.commit`
   - делает commit на текущей ветке.
5. `repo.pr`
   - создает PR (title/body).

## Рекомендуемый ежедневный цикл
1. `repo.snapshot`
2. применить патч
3. `repo.selftest`
4. `repo.diff`
5. `repo.commit`
6. `repo.pr`

## Безопасность
- Никаких секретов в репозитории.
- Все токены только во внешнем хранилище секретов.
- Деструктивные операции (drop/truncate/down -v/rm -rf) запрещены без подтверждения владельца.

## Связь с Opus/Codex orchestration
- Для сложных задач `Claude Opus` используется как planning/review/draft слой.
- Применение изменений в репозиторий и валидация остаются за локальным `Codex`.
- Internal handoff в агентском контуре ведется на English; отчеты пользователю остаются на русском.
- Основной журнал выполнения хранится в существующих `.codex/*` файлах (без нового формата логов).

## UTF-8 and local-only files
- Use explicit UTF-8 for text I/O in PowerShell and shell commands. If output looks garbled, rerun with explicit encoding or use WSL/bash.
- Keep auxiliary local files, downloads, caches, and tool installs under `D:\mudr\_mudro-local` instead of the tracked repo tree.
