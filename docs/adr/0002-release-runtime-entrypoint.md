# ADR 0002: Release Runtime Entrypoint

## Status
Accepted

## Date
2026-04-11

## Context

В репозитории одновременно существуют:

- legacy root `docker-compose.yml`;
- release stack в `docker-compose.prod.yml`;
- dev stack в `ops/compose/`.

Из-за этого разные команды поднимали разные контуры, а `Makefile` частично ссылался на неканонический compose-файл.

## Decision

Для текущего выпуска:

- release stack определяется через `docker-compose.prod.yml`;
- dev/microservice stack определяется через `ops/compose/docker-compose.core.yml` и `ops/compose/docker-compose.services.yml`;
- root `docker-compose.yml` считается legacy/deprecated и не используется как основной entrypoint.

## Consequences

Плюсы:

- один понятный путь для release smoke;
- `Makefile` использует release compose осознанно;
- меньше риска поднять не тот стек.

Минусы:

- legacy root compose пока остаётся в репозитории;
- release stack пока требует отдельного smoke после старта, а не только `depends_on`.

## Follow-up

- либо удалить root `docker-compose.yml`, либо оставить только как явно архивный файл;
- вынести release smoke в отдельную проверку после старта сервисов;
- для showcase path использовать `make health` и `make demo-check` как явные smoke-точки.
