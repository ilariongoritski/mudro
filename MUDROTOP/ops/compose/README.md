# Local compose

Локальный compose-профиль поднимает staging `Postgres` отдельно от основного `mudro`.

## Параметры

- host port: `5434`
- database: `movie_catalog`
- user: `postgres`
- password: `postgres`

## Запуск

```bash
docker compose -f ops/compose/docker-compose.local.yml up -d
```

## Полный локальный профиль

```bash
docker compose -f ops/compose/docker-compose.full.local.yml up --build
```

Этот профиль поднимает:

- `postgres`
- `movie-catalog`
- `frontend`
