---
description: Полный health check mudro для локального и VPS-first контура
---

# Health Check

## Контракт
- Что проверяем: compose, БД, API, frontend/nginx
- Что считаем успехом: `healthz=ok`, БД доступна, frontend отдается
- Риск: read-only, безопасно

## Локальный контур
```powershell
docker compose -f E:\mudr\mudro11-reference\docker-compose.yml ps
docker compose -f E:\mudr\mudro11-reference\docker-compose.yml exec -T db psql -U postgres -d gallery -c "select 1;"
docker compose -f E:\mudr\mudro11-reference\docker-compose.yml exec -T db psql -U postgres -d gallery -c "\dt"
cd E:\mudr\mudro11-reference\frontend && npm.cmd run build
```

## VPS-first контур
```bash
export MUDRO_SERVER_HOST="91.218.113.247"
export MUDRO_SERVER_USER="admin"
export MUDRO_PROJECT_DIR="/srv/mudro"

ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f docker-compose.prod.yml ps"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "curl -fsS http://127.0.0.1:8080/healthz"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "curl -fsS http://127.0.0.1/healthz"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && docker compose -f docker-compose.prod.yml exec -T db psql -U postgres -d gallery -c 'select count(*) from posts;'"
```
