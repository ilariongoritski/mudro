---
name: vps-ssh-ops
description: Безопасное управление VPS-сервером проекта mudro через SSH
---

# Skill: VPS / SSH Operations

Используй этот skill, когда задача относится к VPS runtime, `nginx`, `docker-compose.prod.yml`, логам или security-аудиту.

## Канонические переменные
```bash
export MUDRO_SERVER_HOST="91.218.113.247"
export MUDRO_SERVER_USER="admin"
export MUDRO_PROJECT_DIR="/srv/mudro"
export MUDRO_COMPOSE_FILE="docker-compose.prod.yml"
```

## Режимы

### `audit`
```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f ${MUDRO_COMPOSE_FILE} ps"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "curl -fsS http://127.0.0.1:8080/healthz"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo systemctl status nginx --no-pager"
```

### `runtime`
```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f ${MUDRO_COMPOSE_FILE} ps api agent reporter db"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f ${MUDRO_COMPOSE_FILE} exec -T db psql -U postgres -d gallery -c 'select count(*) from posts;'"
```

### `deploy`
```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && git pull --ff-only origin main"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f ${MUDRO_COMPOSE_FILE} up -d --build"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "curl -fsS http://127.0.0.1:8080/healthz"
```

### `security`
```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo ufw status verbose"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo ss -lntp"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo grep -E '^(PasswordAuthentication|PermitRootLogin)' /etc/ssh/sshd_config"
```

## Роутинг субагентов
- `VPS Auditor`
  - отвечает за `audit` и `runtime`
  - работает только read-only
  - возвращает: статус, точные команды, критерий успеха, замеченные риски
- `Deploy Reviewer`
  - отвечает за `deploy`
  - сначала проверяет последовательность `git pull -> compose up -d --build -> smoke-check`
  - не делает destructive rollback без явного подтверждения
- `Incident Analyst`
  - отвечает за `incident`
  - собирает логи, локализует failure point, предлагает один безопасный retry
  - если retry не помог, обязан эскалировать владельцу
- `Security Reviewer`
  - отвечает за `security`
  - проверяет firewall, sshd, loopback-only DB, отсутствие секретов в tracked tree
  - не меняет `sshd_config`, `ufw`, `iptables` без явного разрешения

## Handoff contract
Каждый субагент обязан вернуть:
1. `Mode`
2. `Goal`
3. `Commands`
4. `Success Criteria`
5. `Risk / Stop Condition`

## Что можно без подтверждения
- read-only статус compose, `nginx`, `healthz`, `df -h`, `free -h`
- логи `docker compose logs` и `journalctl`
- один безопасный retry `docker compose restart api`

## Что требует подтверждения
- изменение `sshd_config`, `ufw`, `iptables`
- изменение SSH-ключей, паролей, panel API key
- `docker compose down -v`
- изменение systemd unit-файлов
- destructive DB-операции

## Ожидаемый контракт ответа
На любой VPS-запрос нужно вернуть:
1. что проверяем или меняем
2. точные команды
3. критерий успеха
4. риск или stop-condition
