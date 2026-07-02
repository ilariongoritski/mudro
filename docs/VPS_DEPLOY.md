# VPS Deploy Guide

## Что подготовлено

| Файл | Назначение |
|------|-----------|
| `.env.production` | Шаблон env-переменных. Заполнить реальными значениями. |
| `ops/scripts/deploy_vps.sh` | Автоматический bootstrap скрипт. |
| `ops/nginx/mudro.conf` | Nginx reverse proxy конфиг. |
| `docker-compose.vps.yml` | Override для VPS (build образов). |
| `Dockerfile.prod` | Multi-stage Go сборка. |
| `frontend/dist/` | Собранный фронтенд (production build). |

## Шаги переезда

### 1. DNS (до VPS)

Добавить A-записи:
```
yourdomain.com       A   <VPS_IP>
api.yourdomain.com   A   <VPS_IP>
```

### 2. На VPS — bootstrap

```bash
# SSH подключение
ssh root@<VPS_IP>

# Клонировать репо
git clone https://github.com/goritskimihail/mudro /opt/mudro
cd /opt/mudro

# Заполнить env
cp .env.production /etc/mudro/.env
nano /etc/mudro/.env   # ← заполнить реальные значения

# Указать домен
export DOMAIN=yourdomain.com

# Запустить
bash ops/scripts/deploy_vps.sh
```

### 3. SSL сертификат

```bash
certbot --nginx -d yourdomain.com -d api.yourdomain.com
```

### 4. BotFather — Mini App

1. @BotFather → /newbot → получить токен
2. /newapp → выбрать бота
3. Имя: `Casino`
4. URL: `https://yourdomain.com/tma/casino`
5. /setmenubutton → URL: `https://yourdomain.com/tma/casino`

### 5. Telegram auth — проверка

TMA откроется в Telegram → initData → `POST /api/auth/telegram` → JWT.

Проверка:
```bash
curl -s https://yourdomain.com/healthz
curl -s https://yourdomain.com/api/casino/balance  # → unauthorized (ok)
```

### 6. OpenClaw (опционально)

```bash
bash ops/scripts/openclaw_gateway_systemd.sh
systemctl enable --now openclaw
```

## Env переменные — обязательные

| Переменная | Описание |
|-----------|----------|
| POSTGRES_PASSWORD | Пароль main DB |
| CASINO_POSTGRES_PASSWORD | Пароль casino DB |
| JWT_SECRET | 32+ chars, рандомный |
| CASINO_INTERNAL_SECRET | Рандомный, для service-to-service |
| TELEGRAM_BOT_TOKEN | От BotFather |
| CASINO_BONUS_TELEGRAM_BOT_TOKEN | Тот же токен |
| OPENROUTER_API_KEY | Для agent LLM |
| DOMAIN | Твой домен |

## Проверка после деплоя

```bash
# Сервисы
docker compose -f docker-compose.prod.yml ps

# Health
curl -s https://yourdomain.com/healthz
curl -s https://yourdomain.com/api/casino/balance

# Frontend
curl -s https://yourdomain.com/ | head -5

# Casino TMA
curl -s https://yourdomain.com/tma/casino | head -5
```

## Бэкапы

```bash
# DB dump
docker compose -f docker-compose.prod.yml exec db pg_dump -U postgres gallery > backup_main.sql
docker compose -f docker-compose.prod.yml exec casino-db pg_dump -U postgres mudro_casino > backup_casino.sql

# Cron (ежедневно 3:00)
echo "0 3 * * * docker compose -f /opt/mudro/docker-compose.prod.yml exec -T db pg_dump -U postgres gallery | gzip > /backup/db_$(date +\%F).sql.gz" | crontab
```
