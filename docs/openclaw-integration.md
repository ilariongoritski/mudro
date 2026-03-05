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
