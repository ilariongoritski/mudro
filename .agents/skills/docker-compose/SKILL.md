---
name: docker-compose-ops
description: Управление Docker Compose сервисами проекта mudro (локально и на VPS)
---

# Skill: Docker Compose Operations

## Конфигурация проекта

| Файл | Среда | Сервисы |
|------|-------|---------|
| `docker-compose.yml` | Локальная разработка | `db` (Postgres 15, порт 5433) |
| `docker-compose.prod.yml` | VPS production | `db`, `api`, `agent`, `reporter` |

## Команды локально (PowerShell)

### Поднять / остановить
```powershell
# Из корня проекта
docker compose up -d          # поднять
docker compose down            # остановить (без удаления данных)
docker compose ps              # статус контейнеров
```

### Логи
```powershell
docker compose logs --tail=100           # последние 100 строк всех сервисов
docker compose logs db --tail=50         # только БД
docker compose logs --follow             # live tail
```

### Рестарт БД
```powershell
docker compose restart db
docker compose exec -T db pg_isready -U postgres  # проверка готовности
```

### Очистка (ОПАСНО — требует подтверждения владельца!)
```powershell
# docker compose down -v   # ⛔ удаляет volume с данными!
```

## Команды на VPS (через SSH)

```bash
cd /srv/mudro
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs api --tail=100
docker compose -f docker-compose.prod.yml restart api
```

## Health Check контейнера БД
```powershell
docker compose exec -T db psql -U postgres -d gallery -c "select 1;"
```
Ожидаемый результат: `1` без ошибок.

## Типичные проблемы

| Проблема | Решение |
|----------|---------|
| `port 5433 already in use` | `docker compose down` или найти процесс: `netstat -ano \| findstr 5433` |
| `Cannot connect to Docker daemon` | Проверить что Docker Desktop запущен |
| Контейнер `unhealthy` | `docker compose logs db --tail=30` и перезапустить |

## Инварианты безопасности
- **Никогда** не запускать `docker compose down -v` без подтверждения
- **Никогда** не менять порт `5433` без обновления всех DSN
- На VPS порт `5432` (внутренний) привязан только к loopback
