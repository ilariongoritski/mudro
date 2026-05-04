# Release Showcase Checklist

## Цель

Подготовить `MUDRO` к внутреннему выпуску и внешнему показу без скрытых переходных сценариев.

## Release path

- auth: backend JWT flow
- runtime: `docker-compose.prod.yml`
- demo navigation: `feed -> chat -> casino -> orchestration`
- showcase helper: `scripts/release-demo.sh`

## Перед стартом

- рабочая ветка без случайного WIP в auth flow;
- `.env` для `docker-compose.prod.yml` заполнен без реальных секретов в git: `MUDRO_APP_DSN`, `CASINO_APP_DSN`, `JWT_SECRET`, MinIO credentials, `CASINO_INTERNAL_SECRET`;
- prod app DSN указывает на non-superuser role (`mudro_app` / `mudro_casino_app`), не на `postgres`;
- API опубликован для nginx как `127.0.0.1:8080`;
- `make selftest` проходит;
- frontend: `npm run lint`, `npm run test`, `npm run build`;
- backend build проходит;
- compose-конфиги валидируются через `docker compose ... config`.

## Smoke after start

1. `GET /healthz` для API
2. `GET /healthz` для casino
3. login / me flow
4. feed non-empty
5. chat page opens after login
6. casino balance available
7. one casino action changes balance/history
8. orchestration page opens without critical error
9. `make demo-check` passes

## Что показываем

- Лента
- Чат
- Казино
- Контур / orchestration

## Что не показываем как release-ready

- незавершённые Supabase traces
- `AuthPage`
- mini app вне Telegram
- admin как production-ready panel без подготовленного admin account

## Known limits

- integration тесты feed пока зависят от `MUDRO_INTEGRATION_TEST_DSN`;
- release compose использует post-start smoke, а не внутренние healthchecks для distroless app containers;
- большой рефакторинг casino store отложен на следующий цикл.
