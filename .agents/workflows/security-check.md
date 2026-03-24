---
description: VPS security check для mudro без изменения состояния
---

# Security Check

## Контракт
- Что проверяем: `ufw`, `sshd`, открытые порты, loopback-only DB, tracked secrets
- Что считаем успехом: наружу доступны только `22/80/443`, `5433` не опубликован, password SSH выключен
- Риск: read-only, безопасно

## VPS проверки
```bash
export MUDRO_SERVER_HOST="91.218.113.247"
export MUDRO_SERVER_USER="admin"

ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo ufw status verbose"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo ss -lntp | grep -E ':22|:80|:443|:5433|:8080' || true"
ssh "${MUDRO_SERVER_USER}@${MUDRO_SERVER_HOST}" "sudo grep -E '^(PasswordAuthentication|PermitRootLogin)' /etc/ssh/sshd_config"
```

## Репозиторные проверки
- убедиться, что ops-скрипты не держат hardcoded credentials
- сверить `docker-compose.prod.yml`, `env/*.env.example`, `docs/vps-control-chat.md`

## Stop condition
Если для фикса требуется менять `sshd_config`, `ufw`, `iptables` или ротировать секреты у провайдера, остановиться и ждать явного подтверждения владельца.
