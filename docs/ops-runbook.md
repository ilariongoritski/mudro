# Ops Runbook

## Цель
Быстрый и детерминированный подъем локального контура `mudro` после ребута/сбоя.

## Предусловия
- рабочая директория: `~/projects/mudro`
- Docker daemon доступен
- Go установлен (`go version`)

## Стандартный recovery (2-3 минуты)
1. Поднять БД:
   - `make up`
2. Проверить контейнер:
   - `docker compose ps`
   - ожидается: `mudro-db-1` в статусе `healthy`
3. Проверить БД:
   - `make dbcheck`
4. Применить миграции:
   - `make migrate`
   - `make migrate-agent`
   - `make migrate-comments`
   - `make migrate-media`
   - `make migrate-comment-model`
5. Проверить таблицы:
   - `make tables`
6. Проверить тесты:
   - `make test`
7. Санити-проверка данных:
   - `make count-posts`

Критерий готовности:
- `dbcheck` OK
- таблицы `posts`, `post_comments`, `post_reactions`, `media_assets`, `comment_reactions`, `agent_queue` существуют
- `make test` проходит

## Политика источников
- `VK` считать архивным snapshot-источником: повторные регулярные обновления VK не планируются.
- Живой импорт и актуализация теперь относятся только к `Telegram`-контуру.

## Частые сбои и действия

### 1) Docker socket permission denied
Симптом:
- `permission denied while trying to connect to the Docker daemon socket ...`

Действия диагностики:
- `id`
- `groups`
- `ls -l /var/run/docker.sock`
- `docker version`

Если запускаешь из ограниченной среды (sandbox/Codex), повторить команды вне sandbox или с разрешенным Docker-доступом.

### 2) Конфликт порта `:8080` у API
Симптом:
- `listen tcp :8080: bind: address already in use`

Действия:
1. `ss -ltnp | grep ':8080' || true`
2. `ps -eo pid,ppid,cmd | grep -E '/tmp/go-build.*/exe/api|go run ./cmd/api' | grep -v grep || true`
3. `kill <PID...>`
4. если не остановились: `kill -9 <PID...>`
5. Проверка: `ps -eo pid,ppid,cmd | grep -E '/tmp/go-build.*/exe/api|go run ./cmd/api' | grep -v grep || true`
6. Перезапуск API: `/usr/local/go/bin/go run ./cmd/api`
7. Проверка: `curl -fsS http://localhost:8080/healthz`

## Ежедневная дисциплина
- минимум 1 осмысленный `commit` + `push` в рабочую ветку
- проверка: `git log --since='today 00:00' --oneline`
