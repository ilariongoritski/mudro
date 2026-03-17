---
description: Полный health check проекта mudro (backend + frontend + БД)
---

# Health Check

Проверка работоспособности всего контура mudro локально.

## Шаги

// turbo-all

1. Проверить что Docker запущен и контейнер БД поднят:
```powershell
docker compose -f d:\mudr\mudro11-main\docker-compose.yml ps
```

2. Если контейнер не запущен, поднять:
```powershell
docker compose -f d:\mudr\mudro11-main\docker-compose.yml up -d
```

3. Проверить подключение к БД:
```powershell
docker compose -f d:\mudr\mudro11-main\docker-compose.yml exec -T db psql -U postgres -d gallery -c "select 1;"
```

4. Проверить наличие таблиц:
```powershell
docker compose -f d:\mudr\mudro11-main\docker-compose.yml exec -T db psql -U postgres -d gallery -c "\dt"
```

5. Проверить количество постов:
```powershell
docker compose -f d:\mudr\mudro11-main\docker-compose.yml exec -T db psql -U postgres -d gallery -c "select source, count(*) from posts group by source;"
```

6. Проверить фронтенд сборку:
```powershell
cd d:\mudr\mudro11-main\frontend && npm.cmd run build
```

## Критерии успеха
- Docker контейнер `db` в статусе `healthy`
- `select 1` = ok
- Таблица `posts` существует
- `npm run build` завершается без ошибок
