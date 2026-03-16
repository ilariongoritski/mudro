# Лог прогона 20260316-1109

## Окружение
- `pwd`: `D:\mudr\mudro11`
- Ветка: `codex/frontend-mudro11-fsd`
- Базовый коммит до фиксации: `569d1d6`
- Цель: довести серверный контур до базово стабильного состояния без публичного Postgres и без падений `mudro-bot`/`mudro-reporter`.

## Что проверял
1. Состояние VPS Postgres и systemd-сервисов через удаленный shell.
2. Конфиг `mudro-api.service`, `.env` и docker compose bind для БД.
3. Фактические ошибки в логах Postgres и `mudro-bot.service`.
4. Локальный health loop по затронутым пакетам.
5. Внешний health публичного preview и VPS API.

## Ключевые наблюдения
- Postgres на VPS был выставлен наружу на `0.0.0.0:5433`.
- `mudro-api.service` и серверный `.env` использовали DSN с суперпользователем `postgres`.
- В логах БД были реальные внешние попытки auth/SQL abuse; это объяснило периодические `failed SASL auth`.
- `mudro-bot.service` падал из-за stale-файла `cmd/bot/gototg.go`, которого нет в актуальном дереве репы.
- `mudro-reporter.service` падал из-за отсутствующего `REPORT_CHAT_ID`.

## Что изменил в репозитории
- `docker-compose.yml`
  - bind БД переведен с `5433:5432` на `127.0.0.1:${POSTGRES_PORT:-5433}:5432`
  - пароль БД вынесен в `${POSTGRES_PASSWORD:-postgres}`
- `docker-compose.prod.yml`
  - bind db/redis/kafka ограничен loopback
- `env/db.env.example`
  - добавлена явная пометка про сильный пароль и запрет на дефолтный `postgres`
- `README.md`
  - добавлено предупреждение про VPS hardening Postgres
- `docs/ops-runbook.md`
  - добавлен раздел hardening Postgres на VPS
- `scripts/ops/harden_vps_db_auth.sh`
  - создает `mudro_app`
  - меняет DSN на `mudro_app`
  - ротирует пароль суперпользователя
  - включает systemd drop-in и host firewall guard для `5433`

## Что выполнил на VPS
1. Синхронизировал измененные файлы в `/root/projects/mudro`.
2. Запустил `scripts/ops/harden_vps_db_auth.sh` с секретами из локального вне-репозиторного env.
3. Проверил:
   - `mudro-api.service`
   - `mudro-db-firewall.service`
   - `ss -lntp | grep 5433`
   - `iptables -S INPUT | grep 5433`
4. Досинхронизировал серверный bot/reporter-контур:
   - `cmd/bot`
   - `cmd/reporter`
   - `internal/bot`
   - `internal/reporter`
   - `internal/config`
   - `go.mod`
   - `go.sum`
5. Удалил stale `.go`-артефакты из `/root/projects/mudro/cmd/bot`, кроме `main.go`.
6. Прописал `REPORT_CHAT_ID=362151731` в серверный `.env`.
7. Перезапустил `mudro-bot.service` и `mudro-reporter.service`.

## Проверки
### Локально
- `make dbcheck` -> ok
- `go test ./internal/api ./internal/commentmodel ./cmd/commentbackfill ./cmd/mediabackfill` -> ok

### VPS
- `systemctl is-active mudro-api.service mudro-bot.service mudro-reporter.service mudro-db-firewall.service` -> все `active`
- `ss -lntp | grep 5433` -> только `127.0.0.1:5433`
- `iptables -S INPUT | grep 5433` -> есть DROP правило для внешнего `tcp/5433`
- `journalctl -u mudro-bot -n 5` -> `Авторизован как Mudrot_bot`
- `journalctl -u mudro-reporter -n 5` -> reporter авторизован и видит `chat_id=362151731`
- `docker logs --since 1m mudro-db-1 | grep 'password authentication failed|failed SASL auth'` -> пусто после hardening
- `curl -fsS http://127.0.0.1:8080/healthz` -> `{"status":"ok"}`
- `curl -fsS http://127.0.0.1:8080/api/front?source=tg&limit=1` -> valid JSON
- `curl -fsS https://frontend-psi-ten-33.vercel.app/healthz` -> `{"status":"ok"}`

## Ретраи и проблемы
- Первый remote-run не передал `MUDRO_DB_APP_PASSWORD` из-за пустой PowerShell-подстановки; повтор выполнен с явной подстановкой секрета.
- `mudro-bot.service` падал до очистки stale-файлов; после выравнивания tree сервис поднялся.

## Итог
- Корень нестабильности был инфраструктурный: публичный `5433` + суперпользователь `postgres` в DSN.
- Серверный контур переведен на `mudro_app`, DB bind ограничен loopback, внешний `5433` режется host firewall.
- `mudro-api`, `mudro-bot`, `mudro-reporter` одновременно живы.
- Следующий внешний шаг: убрать публичный `:8080` за reverse proxy/HTTPS.
