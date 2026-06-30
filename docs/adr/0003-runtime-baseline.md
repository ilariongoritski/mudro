# ADR 0003: Runtime Baseline Matrix

## Status
Accepted

## Date
2026-06-29

## Context

В репозитории одновременно существуют:

- product-facing MVP runtime: `feed-api`, `casino`, `frontend`, `agent`, `bot`;
- additive сервисы: `auth-api`, `api-gateway`, `bff-web`, `orchestration-api`, `movie-catalog`;
- transitional контуры: root `docker-compose.yml`, старые страницы и старые recovery-path команды.

Из-за этого один и тот же проект запускался разными наборами сервисов и разными bootstrap-командами. Главная проблема была не в коде, а в расхождении канона между `Makefile`, compose, runbook и вспомогательными скриптами.

## Decision

Для текущего pre-MVP baseline:

- локальный core runtime определяется через:
  - `ops/compose/docker-compose.core.yml`
  - `ops/compose/docker-compose.casino.local.yml`
  - `make health`
  - `make core-up`
  - `make dbcheck-core`
  - `make migrate-runtime`
- production-style runtime определяется через:
  - `docker-compose.prod.yml`
  - nginx/systemd как deploy-specific overlay, а не как отдельная модель продукта;
- активные продуктовые runtime-сервисы:
  - `services/feed-api`
  - `services/casino`
  - `services/agent`
  - `services/bot`
- additive, но не core:
  - `services/auth-api`
  - `services/api-gateway`
  - `services/bff-web`
  - `services/orchestration-api`
  - `services/movie-catalog`
- transitional:
  - root `docker-compose.yml`
  - root `cmd/mudro` оставлен как канонический агрегирующий CLI
  - устаревшие frontend entrypoints, которые не участвуют в каноничном router path.

## Consequences

Плюсы:

- один локальный bootstrap path для новой работы;
- `Makefile`, worker loop и Windows helpers больше не учат legacy recovery-потоку;
- migration inventory считается полным только если включает main schema и `movie_catalog`.

Минусы:

- `feed-api` остаётся modular monolith и ещё не отражает будущий service split;
- `systemd + nginx` всё ещё нужны для VPS rollout, но считаются deploy overlay поверх runtime, а не отдельной продуктовой моделью;
- additive runtime остаётся в репозитории и требует дисциплины, чтобы не восприниматься как обязательный core.

## Follow-up

- отдельно принять решение по `feed-api`: modular monolith vs реальный split;
- продолжить разрез `services/casino/internal/casino/store.go`;
- убрать или изолировать transitional frontend/backend entrypoints, которые выглядят активными, но не входят в канон.
