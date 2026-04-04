---
description: Деплой frontend на VPS через server-side rollout script
---

# Deploy Frontend

## Контракт
- Что делаем: собираем `frontend/dist`, затем обновляем `nginx`-контур на VPS
- Что считаем успехом: frontend раздается из `/var/www/mudro/frontend`, `/healthz` отвечает
- Риск: краткий даунтайм во время restart `nginx`

## Шаги
1. Собрать бандл:
```powershell
cd E:\mudr\mudro11-reference\frontend
npm.cmd ci
npm.cmd run build
```

2. На VPS выполнить server-side rollout:
```bash
export MUDRO_SERVER_HOST="91.218.113.247"
export MUDRO_SERVER_USER="admin"
export MUDRO_PROJECT_DIR="/srv/mudro"

ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "cd ${MUDRO_PROJECT_DIR} && sudo bash scripts/ops/deploy_vps_frontend.sh"
```

3. Smoke-check:
```bash
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "curl -fsS http://127.0.0.1/healthz"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "curl -I http://127.0.0.1/"
```
