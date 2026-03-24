---
name: docker-compose-ops
description: Управление Docker Compose сервисами проекта mudro (локально и на VPS)
---

# Skill: Docker Compose Operations

## Конфигурация проекта

| Файл | Среда | Сервисы |
|------|-------|---------|
| `docker-compose.yml` | локальная разработка | `db`, `casino-db`, `casino-api` |
| `docker-compose.prod.yml` | VPS-first runtime | `db`, `casino-db`, `casino-api`, `redis`, `kafka`, `api`, `agent`, `reporter`, `minio` |

## Prod runtime
- проект на VPS: `/srv/mudro`
- прод-контур читает секреты только из локальных `env/*.env`
- `api` публикуется на `127.0.0.1:8080`
- `db` публикуется только на `127.0.0.1:5433`

## Команды локально
```powershell
docker compose up -d
docker compose ps
docker compose logs --tail=100
docker compose exec -T db psql -U postgres -d gallery -c "select 1;"
```

## Команды на VPS
```bash
cd /srv/mudro
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs api --tail=100
docker compose -f docker-compose.prod.yml restart api
docker compose -f docker-compose.prod.yml exec -T db psql -U postgres -d gallery -c "select 1;"
```

## Типичные проблемы
| Проблема | Действие |
|----------|----------|
| `env file ... not found` | создать отсутствующий `env/*.env` из `*.env.example` |
| `Cannot connect to Docker daemon` | проверить Docker daemon и права пользователя |
| контейнер `unhealthy` | посмотреть `logs`, затем сделать один безопасный restart |

## Инварианты безопасности
- никогда не запускать `docker compose down -v` без подтверждения
- не хранить секреты в `docker-compose.prod.yml`
- на VPS использовать только `env/*.env`, а не tracked `.env`
