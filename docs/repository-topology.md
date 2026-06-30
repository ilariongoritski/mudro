# Repository Topology

## Active Zones

- `services/` — long-running runtime services.
- `tools/` — importers, backfill, maintenance CLI.
- `internal/` — shared domain/application logic.
- `contracts/` — service interfaces (HTTP + events).
- `ops/` — compose, runbooks, scripts, env docs.
- `platform/agent-control/` — unified agent governance/policies/profiles.
- `frontend/` — React/TypeScript UI.
- `migrations/` — SQL schema changes.

## Compose Entrypoints

- `ops/compose/docker-compose.core.yml` — основной локальный runtime.
- `ops/compose/docker-compose.casino.local.yml` — отдельный локальный casino runtime.
- `docker-compose.prod.yml` — production-style VPS/self-hosted runtime.

## CLI Entrypoints

- `cmd/mudro` — канонический агрегирующий CLI.
- `tools/` — одноразовые import/backfill/maintenance команды.

## Removed Legacy

- Root thin-wrapper `cmd/*` и `legacy/old/*` удалены из active tree.
- Удаленные legacy entrypoints не должны появляться в Makefile/CI/runbook как рабочие команды.
