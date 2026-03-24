# Ops Runbook

## Цель
Быстрый и детерминированный подъем локального и VPS-first контура `mudro` после ребута, сбоя или deploy.

## Источники истины
- `docs/vps-control-chat.md` — контракт управляющего чата
- `docker-compose.prod.yml` — основной VPS runtime
- `scripts/ops/*` — прикладные ops-скрипты
- `Makefile` — локальный и сервисный baseline

## Канонические пути
- локально: `~/projects/mudro`
- на VPS: `/srv/mudro`

## Локальный recovery (2-3 минуты)
1. Поднять локальный контур:
   - `make health`
2. Если нужен пошаговый прогон:
   - `make up`
   - `docker compose ps`
   - `make dbcheck`
   - `make migrate`
   - `make tables`
   - `make test`
   - `make count-posts`

Критерий готовности:
- `dbcheck` OK
- таблицы `posts`, `post_comments`, `media_assets`, `agent_queue` существуют
- `make test` проходит

## VPS-First Runtime
Основной серверный контур:
- `docker-compose.prod.yml`
- `nginx` раздает frontend из `/var/www/mudro/frontend`
- `nginx` проксирует `/api`, `/media`, `/healthz` на `127.0.0.1:8080`
- Postgres опубликован только на `127.0.0.1:5433`

Базовые проверки на VPS:

```bash
cd /srv/mudro
docker compose -f docker-compose.prod.yml ps
curl -fsS http://127.0.0.1:8080/healthz
sudo systemctl status nginx --no-pager
```

## Frontend на VPS через nginx
Разовый rollout:
1. собрать frontend:
   - `cd frontend`
   - `npm.cmd run build`
2. на VPS запустить:
   - `bash /srv/mudro/scripts/ops/deploy_vps_frontend.sh`
3. проверить:
   - `curl -fsS http://127.0.0.1/healthz`
   - `curl -I http://127.0.0.1/`
   - `sudo ss -lntp | grep ':80'`

Результат:
- сайт открывается с VPS
- `nginx` стал основной публичной точкой входа
- Vercel больше не обязателен как runtime

## Hardening Postgres на VPS
Цель: не держать публично доступный `postgres/postgres` и не использовать `.env` в tracked workflow.

Базовый безопасный контур:
- БД слушает только `127.0.0.1:5433`
- приложение ходит под отдельным `mudro_app`
- пароль `postgres` хранится только в `env/db.env` на VPS
- входящий `tcp/5433` режется вне loopback

Разовый шаг на VPS:
1. задать секреты в shell:
   - `export MUDRO_DB_APP_PASSWORD='<strong password>'`
   - `export MUDRO_DB_SUPERUSER_PASSWORD='<strong password>'`
2. запустить:
   - `sudo -E bash scripts/ops/harden_vps_db_auth.sh`
3. проверить:
   - `docker compose -f docker-compose.prod.yml ps db api agent`
   - `curl -fsS http://127.0.0.1:8080/healthz`
   - `sudo ss -lntp | grep 5433`

Ожидаемый результат:
- `api` и `agent` поднимаются на `env/*.env`
- `5433` слушает только loopback
- внешний auth-шум по Postgres исчезает

## Частые сбои и действия

### Docker socket permission denied
Симптом:
- `permission denied while trying to connect to the Docker daemon socket ...`

Диагностика:
- `id`
- `groups`
- `ls -l /var/run/docker.sock`
- `docker version`

### Конфликт порта `:8080`
Симптом:
- `listen tcp :8080: bind: address already in use`

Действия:
1. `sudo ss -ltnp | grep ':8080' || true`
2. `docker compose -f docker-compose.prod.yml ps`
3. `docker compose -f docker-compose.prod.yml logs api --tail=100`
4. безопасный retry:
   - `docker compose -f docker-compose.prod.yml restart api`

### Frontend отвечает, API нет
Проверить:
- `curl -fsS http://127.0.0.1:8080/healthz`
- `sudo systemctl status nginx --no-pager`
- `docker compose -f docker-compose.prod.yml logs api --tail=100`

## Ежедневная дисциплина
- минимум один осмысленный `commit` и `push` в рабочую ветку
- перед deploy: `git pull --ff-only`
- после deploy: `docker compose -f docker-compose.prod.yml ps` и `curl -fsS http://127.0.0.1:8080/healthz`
