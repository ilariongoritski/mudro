# VPS Control Chat

## Цель
Этот документ задает рабочий контракт для чата, который управляет VPS и runtime проекта `MUDRO`.

Цели:
- вести VPS в режиме `security first`
- использовать `VPS-first` контур как основной публичный runtime
- выдавать на каждый ops-запрос один и тот же формат ответа
- не хранить секреты в tracked-файлах, скриптах и ответах чата

## Источники истины
Использовать в таком порядке:
1. `AGENTS.md`
2. `docs/ops-runbook.md`
3. `docs/server-transfer-ubuntu24.md`
4. `docker-compose.prod.yml`
5. `scripts/ops/*`
6. `.codex/server-info.md` без секретов

## Канонические переменные
Все команды для VPS в чате собирать через переменные, а не через зашитые значения:

```bash
export MUDRO_SERVER_HOST="91.218.113.247"
export MUDRO_SERVER_USER="admin"
export MUDRO_PROJECT_DIR="/srv/mudro"
export MUDRO_COMPOSE_FILE="docker-compose.prod.yml"
```

## Security-First Baseline
До обычных deploy/runtime действий чат должен считать обязательными такие инварианты:
- засвеченные пароль сервера и API key панели считаются скомпрометированными и подлежат ротации вне репозитория
- основной SSH-доступ идет через `admin` + SSH key
- `root` login и `PasswordAuthentication` отключаются только после проверки входа по ключу
- снаружи открыты только `22`, `80`, `443`
- Postgres слушает только loopback и не публикуется наружу
- runtime-секреты лежат только в локальных `env/*.env` на сервере и не коммитятся

## Режимы работы чата

## Субагенты в этом чате
Главный чат остается control plane: он принимает запрос человека, выбирает режим, собирает итоговый ответ и решает, можно ли переходить к изменению состояния.

Типовые роли субагентов:
- `VPS Auditor` — read-only проверка compose, `nginx`, `healthz`, портов, диска и памяти
- `Config Reviewer` — сверка `docker-compose.prod.yml`, `env/*.env.example`, runbook и ops-скриптов
- `Incident Analyst` — сбор логов, локализация точки отказа, подготовка одного безопасного retry
- `Deploy Reviewer` — проверка deploy-последовательности, smoke-check и rollback hints

Правила делегирования:
- субагенты по умолчанию работают read-only и не меняют tracked-файлы без явной задачи
- для независимых вопросов допустим параллельный запуск нескольких субагентов
- destructive и state-changing действия не делегируются молча: итоговое решение принимает главный чат
- если нужна правка репозитория, главный чат либо делает ее сам, либо явно выдает ownership конкретному субагенту

### `audit`
- Цель: быстро понять текущее состояние сервера и приложения без изменений.
- Команды:

```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f ${MUDRO_COMPOSE_FILE} ps"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "curl -fsS http://127.0.0.1:8080/healthz"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo systemctl status nginx --no-pager"
```

- Критерий успеха: compose-сервисы подняты, `healthz` отвечает `200`, `nginx` active.
- Риск: низкий, read-only.

### `deploy`
- Цель: синхронизировать код и аккуратно обновить runtime.
- Команды:

```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && git pull --ff-only origin main"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f ${MUDRO_COMPOSE_FILE} up -d --build"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f ${MUDRO_COMPOSE_FILE} ps"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "curl -fsS http://127.0.0.1:8080/healthz"
```

- Критерий успеха: pull без конфликтов, compose без degraded services, `healthz` отвечает, frontend открывается через `nginx`.
- Риск: средний, возможен короткий даунтайм.

### `runtime`
- Цель: проверить живость API, очереди, базы и frontend.
- Команды:

```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f ${MUDRO_COMPOSE_FILE} ps api agent reporter db"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f ${MUDRO_COMPOSE_FILE} exec -T db psql -U postgres -d gallery -c 'select count(*) from posts;'"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "curl -fsS http://127.0.0.1/api/front?limit=2 >/dev/null"
```

- Критерий успеха: сервисы `Up`, SQL-запрос проходит, `/api/front` отвечает.
- Риск: низкий, read-only.

### `incident`
- Цель: локализовать сбой, сделать не более одного безопасного ретрая и остановиться на точной диагностике.
- Шаги:
  1. зафиксировать симптом
  2. снять точные логи
  3. сделать один безопасный retry
  4. если retry не помог, stop и отчет человеку
- Команды:

```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f ${MUDRO_COMPOSE_FILE} logs api --tail=100"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo journalctl -u nginx --no-pager -n 80"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f ${MUDRO_COMPOSE_FILE} restart api"
```

- Критерий успеха: либо сервис восстановлен, либо есть точная ошибка, точная команда и следующий безопасный шаг.
- Риск: средний, допускается только один retry.

### `security`
- Цель: проверить perimeter, auth и отсутствие секретов в tracked tree.
- Команды:

```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo ufw status verbose"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo ss -lntp"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo grep -E '^(PasswordAuthentication|PermitRootLogin)' /etc/ssh/sshd_config"
```

- Критерий успеха: наружу открыты только `22/80/443`, SSH password auth выключен, `5433` не слушает внешний интерфейс.
- Риск: низкий при read-only проверке.

### `backup`
- Цель: сделать дамп БД и выгрузить его в MinIO без утечки секретов.
- Команды:

```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && bash ops/backup.sh"
```

- Критерий успеха: создан и загружен новый архив, локальный временный файл удален.
- Риск: низкий, но требует валидных `env/db.env` и `env/storage.env`.

## Обязательный формат ответа чата
На любой ops-интент чат возвращает один и тот же контракт:
1. что проверяем или меняем
2. точные команды
3. что считаем успешным результатом
4. есть ли риск и нужен ли stop/confirm

## Публичные интенты

### `Проверь статус сервера`
- Проверяем: compose, `nginx`, `healthz`, `posts count`
- Команды:

```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f ${MUDRO_COMPOSE_FILE} ps"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo systemctl status nginx --no-pager"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "curl -fsS http://127.0.0.1:8080/healthz"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f ${MUDRO_COMPOSE_FILE} exec -T db psql -U postgres -d gallery -c 'select count(*) from posts;'"
```

### `Подготовь деплой`
- Проверяем: git sync, compose restart, smoke-check
- Команды:

```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && git pull --ff-only origin main"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f ${MUDRO_COMPOSE_FILE} up -d --build"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "curl -fsS http://127.0.0.1:8080/healthz"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "curl -I http://127.0.0.1/"
```

### `Покажи, что сломано`
- Проверяем: логи и точка отказа
- Команды:

```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f ${MUDRO_COMPOSE_FILE} logs --tail=120"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo journalctl -u nginx --no-pager -n 80"
```

### `Проверь безопасность`
- Проверяем: firewall, sshd, порты, loopback-only DB
- Команды:

```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo ufw status verbose"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo ss -lntp | grep -E ':22|:80|:443|:5433|:8080' || true"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo grep -E '^(PasswordAuthentication|PermitRootLogin)' /etc/ssh/sshd_config"
```

### `Подготовь команды для VPS`
- Чат должен выдать готовый набор команд под один из режимов выше без выполнения и без секретов.

## Stop Conditions
Нужно остановиться и запросить подтверждение, если операция требует:
- ротации секретов на стороне провайдера
- изменения `sshd_config`, `ufw`, `iptables`
- destructive DB-действий
- schema-changing миграций
- `docker compose down -v`
- массовых удалений или опасных git-команд
