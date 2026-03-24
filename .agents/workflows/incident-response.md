---
description: Incident triage для VPS-first runtime mudro
---

# Incident Response

## Контракт
- Что делаем: локализуем сбой, собираем точные логи, допускаем один безопасный retry
- Что считаем успехом: либо сервис восстановлен, либо есть точная причина и следующий безопасный шаг
- Риск: средний, один retry максимум

## Шаги
1. Зафиксировать симптом пользователя.
2. Снять compose и systemd логи:
```bash
export MUDRO_SERVER_HOST="91.218.113.247"
export MUDRO_SERVER_USER="admin"
export MUDRO_PROJECT_DIR="/srv/mudro"

ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f docker-compose.prod.yml logs --tail=120"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo journalctl -u nginx --no-pager -n 80"
```
3. Проверить health endpoint:
```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "curl -fsS http://127.0.0.1:8080/healthz"
```
4. Если причина похожа на временный runtime-сбой, сделать один безопасный retry:
```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f docker-compose.prod.yml restart api"
```

## Stop condition
- не делать второй retry автоматически
- не выполнять `docker compose down -v`
- не менять схему БД, firewall, sshd или секреты без подтверждения владельца
