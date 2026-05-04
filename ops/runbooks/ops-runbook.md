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
   - `make migrate-list` для проверки порядка up-миграций
   - `make check-migration-up-list` перед bootstrap/recovery, если менялся migration inventory
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

### Frontend на VPS через nginx
Цель: держать основной MVP frontend прямо на VPS, а не только на внешнем preview-хостинге.

Ожидаемая схема:
- `nginx` слушает `:80`
- статика frontend лежит в `/var/www/mudro/frontend`
- `nginx` проксирует `/api`, `/media`, `/healthz` на `127.0.0.1:8080`

Разовый rollout:
1. локально собрать frontend:
   - `cd frontend`
   - `npm.cmd run build`
2. загрузить проект на VPS или хотя бы актуальный `frontend/dist`
3. на VPS запустить:
   - `bash /root/projects/mudro/ops/scripts/deploy_vps_frontend.sh`
4. проверить:
   - `curl -fsS http://127.0.0.1/healthz`
   - `curl -I http://127.0.0.1/`
   - `ss -lntp | grep ':80'`

Результат:
- сайт открывается по `http://<server-ip>/`
- API и media продолжают жить на том же VPS, но уже за reverse proxy
- Vercel перестает быть обязательной точкой входа для MVP

HTTPS rollout, Certbot, security headers and external smoke checklist:
- `ops/runbooks/vps-https-nginx.md`

Secrets rotation without writing secrets to the repo:
- `ops/runbooks/secrets-rotation.md`

### 0) Hardening Postgres на VPS
Цель: не держать публично доступный `postgres/postgres` на `0.0.0.0:5433`.

Базовый безопасный контур:
- БД снаружи слушает только `127.0.0.1:5433`
- сервисы приложения ходят не под `postgres`, а под отдельным `mudro_app`
- пароль `postgres` на VPS не должен оставаться дефолтным
- хост дополнительно режет входящий `tcp/5433` вне loopback через systemd-managed `iptables` rule

Разовый шаг на VPS:
1. задать секреты в shell:
   - `export MUDRO_DB_APP_PASSWORD='<strong password>'`
   - `export MUDRO_DB_SUPERUSER_PASSWORD='<strong password>'`
2. запустить:
   - `bash ops/scripts/harden_vps_db_auth.sh`
3. проверить:
   - `systemctl status mudro-api --no-pager`
   - `curl -fsS http://127.0.0.1:8080/healthz`
   - `ss -lntp | grep 5433`

Ожидаемый результат:
- `mudro-api` работает
- `5433` слушает только `127.0.0.1`
- в journal Postgres больше нет внешнего auth-шума по публичному порту

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
2. `ps -eo pid,ppid,cmd | grep -E '/tmp/go-build.*/exe/api|go run ./services/feed-api/cmd' | grep -v grep || true`
3. `kill <PID...>`
4. если не остановились: `kill -9 <PID...>`
5. Проверка: `ps -eo pid,ppid,cmd | grep -E '/tmp/go-build.*/exe/api|go run ./services/feed-api/cmd' | grep -v grep || true`
6. Перезапуск API: `/usr/local/go/bin/go run ./services/feed-api/cmd`
7. Проверка: `curl -fsS http://localhost:8080/healthz`

## Ежедневная дисциплина
- минимум 1 осмысленный `commit` + `push` в рабочую ветку
- проверка: `git log --since='today 00:00' --oneline`

## Runtime Bootstrap (P0)

Canonical runtime checks now use active core compose and full runtime migrations:

```bash
make core-up
make core-ps
make dbcheck-core
make migrate-list
make migrate-runtime
make tables-core
make test-active
make count-posts-core
```

One-command health loop:

```bash
make health
```

## Production compose hardening

`docker-compose.prod.yml` is the production-style smoke stack. It must not run API, agent, or casino services with `postgres` superuser DSN. Telegram/report bots are kept outside the default production stack and should be started as a separate, explicitly approved contour.

Required `.env` values before `docker compose -f docker-compose.prod.yml up -d`:

Required variables for the production shell or secret manager:

```text
POSTGRES_PASSWORD
CASINO_POSTGRES_PASSWORD
MUDRO_APP_DSN
CASINO_APP_DSN
JWT_SECRET
CASINO_INTERNAL_SECRET
MINIO_ROOT_USER
MINIO_ROOT_PASSWORD
```

Fail-fast guards are intentional: missing `JWT_SECRET`, MinIO credentials, `CASINO_INTERNAL_SECRET`, or app DSNs should stop compose config/start before containers run with weak defaults.

Before switching services to app DSNs, create/grant non-superuser roles. Main DB can use the existing VPS helper:

```bash
export MUDRO_DB_APP_PASSWORD='<app password>'
export MUDRO_DB_SUPERUSER_PASSWORD='<postgres superuser password>'
bash ops/scripts/harden_vps_db_auth.sh
```

For the casino DB, create an equivalent non-superuser role and grants:

```sql
CREATE ROLE mudro_casino_app LOGIN PASSWORD '<casino app password>' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
GRANT CONNECT ON DATABASE mudro_casino TO mudro_casino_app;
GRANT USAGE ON SCHEMA public TO mudro_casino_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO mudro_casino_app;
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO mudro_casino_app;
ALTER DEFAULT PRIVILEGES FOR USER postgres IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO mudro_casino_app;
ALTER DEFAULT PRIVILEGES FOR USER postgres IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO mudro_casino_app;
```

API port canonical binding for nginx is `127.0.0.1:8080`. Keep `API_PORT` unset unless there is an explicit local conflict; nginx should proxy `/api`, `/media`, and `/healthz` to `127.0.0.1:8080`.

## Casino on Supabase

Use this when the casino contour should run against Supabase instead of the local `casino-db` container.

Checklist:
- set `CASINO_DSN` to the Supabase direct connection string and keep `sslmode=require`
- keep `CASINO_START_BALANCE=500`
- inspect order with `make migrate-casino-list` or `make migrate-casino-dry-run`
- run `bash ./scripts/migrate-casino.sh`
- verify the casino service health endpoint after migration

Notes:
- `scripts/migrate-casino.sh` applies only up migrations from `services/casino/migrations/` and excludes `*.down.sql`
- legacy slot tables stay in place; the new casino tables extend the schema for multi-game history and roulette

## Railway rollout

Railway-specific split deployment for `frontend + feed-api + casino` is described in:

- `ops/runbooks/railway-casino-miniapp.md`

Use that runbook when the mini app should be deployed as three Railway services with `casino` pointed at Supabase.

## Local Demo (localhost, no Vercel)

```bash
make demo-up
npm.cmd --prefix frontend run dev
make demo-check
```

Expected endpoints:
- `http://127.0.0.1:8080/healthz`
- `http://127.0.0.1:8082/healthz`
- `http://127.0.0.1:5173`

Notes:
- `make health` now covers both the core runtime and the separate casino DB/API contour.
- `make demo-up` applies runtime migrations and auto-seeds the demo feed from `data/nu/feed_items.json` if the local `posts` table is still empty.
- `make demo-check` now validates both API health and that `/api/front` returns a non-empty feed.
