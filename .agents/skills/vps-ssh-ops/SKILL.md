---
name: vps-ssh-ops
description: Безопасное управление VPS-сервером проекта mudro через SSH
---

# Skill: VPS / SSH Operations

## Параметры сервера
- **IP:** `91.218.113.247`
- **OS:** Ubuntu 24.04
- **Рабочая копия:** `/srv/mudro`
- **Сервисы (systemd):** `mudro-api`, `mudro-bot`, `mudro-reporter` (может быть выключен)
- **Пользователь для SSH:** `admin` (или `root` с ключом)

## Безопасные операции (без подтверждения)

### Проверка статуса сервисов
```bash
ssh admin@91.218.113.247 "systemctl status mudro-api mudro-bot --no-pager"
```

### Логи сервисов
```bash
ssh admin@91.218.113.247 "journalctl -u mudro-api --no-pager -n 50"
ssh admin@91.218.113.247 "journalctl -u mudro-bot --no-pager -n 30"
```

### Health check API
```bash
ssh admin@91.218.113.247 "curl -s http://127.0.0.1:8080/healthz"
```
Ожидаемый ответ: `{"status":"ok"}`

### Диск и память
```bash
ssh admin@91.218.113.247 "df -h / && free -h"
```

### Docker на сервере
```bash
ssh admin@91.218.113.247 "docker ps --format 'table {{.Names}}\t{{.Status}}'"
```

### UFW статус
```bash
ssh admin@91.218.113.247 "sudo ufw status verbose"
```

## Операции, требующие осторожности

### Рестарт API (безопасно, но вызывает даунтайм ~2с)
```bash
ssh admin@91.218.113.247 "sudo systemctl restart mudro-api"
```

### Обновить код на сервере (git pull)
```bash
ssh admin@91.218.113.247 "cd /srv/mudro && git pull origin main"
```

### Пересобрать и перезапустить API
```bash
ssh admin@91.218.113.247 "cd /srv/mudro && go build -o /usr/local/bin/mudro-api ./cmd/api && sudo systemctl restart mudro-api"
```

## ⛔ Запрещено без подтверждения владельца
- Изменение `sshd_config`, UFW правил, `iptables`
- Удаление файлов из `/srv`, `/etc`, `/var/lib`, `/root`, `/home`
- `docker compose down -v`
- Изменение паролей и SSH-ключей
- Изменение systemd unit-файлов

## SSH из PowerShell (Windows)
```powershell
# Не использовать WSL-ключи из /mnt/c (ошибка прав 0777)
# Использовать нативный PowerShell ssh:
ssh -i $env:USERPROFILE\.ssh\id_rsa admin@91.218.113.247 "команда"
```

## Типичные проблемы

| Проблема | Диагностика | Решение |
|----------|-------------|---------|
| API не отвечает | `systemctl status mudro-api` | `sudo systemctl restart mudro-api` |
| БД недоступна | `docker ps`, `pg_isready` | `docker compose up -d db` |
| Диск заполнен | `df -h /`, `du -sh /var/log/*` | Очистить логи Docker: `docker system prune` |
| SSH таймаут | Проверить UFW: `ufw status` | Убедиться что порт 22 открыт |
