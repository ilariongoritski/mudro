# MUDRO

MUDRO is a microservices-first monorepo for feed ingestion, agent workflows, Telegram operations, casino runtime, frontend UI, and data import tooling.

## Release / Showcase State

Current release line: `v0.1.0-mvp`.

The current showcase is centered on a single clear path:
- backend auth flow only
- canonical local runtime via `ops/compose/docker-compose.core.yml`
- release smoke via `docker-compose.prod.yml`
- separate casino contour with its own DB and wallet sync

## Canonical Runtime

Use these contours deliberately:

- `ops/compose/docker-compose.core.yml` — canonical local runtime for `db`, `redis`, `kafka/redpanda`, `api`, and `agent`
- `ops/compose/docker-compose.services.yml` — additive service layer for `auth-api`, `orchestration-api`, `bff-web`, and `api-gateway`
- `docker-compose.prod.yml` — release stack for production-style smoke
- `docker-compose.yml` — legacy root contour, kept only for compatibility and recovery

## Quick Start

Canonical local bootstrap:

```bash
make core-up
make dbcheck-core
make migrate-runtime
make tables-core
make health-runtime
```

Full MVP bootstrap, including the separate casino database/runtime:

```bash
make health
```

Showcase flow:

```bash
make demo-up
cd frontend && npm.cmd run dev
make demo-check
```

Or use the helper script:

```bash
bash ./scripts/release-demo.sh
```

Expected endpoints:
- API health: `http://127.0.0.1:8080/healthz`
- Casino health: `http://127.0.0.1:8082/healthz`
- Frontend dev server: `http://127.0.0.1:5173`

## Casino Contour

Casino is intentionally separate from the main runtime.

For the local casino DB:
- set `CASINO_DSN` to the casino database
- set `CASINO_MAIN_DSN` to the main Mudro wallet DB
- keep `CASINO_START_BALANCE=500`
- apply both migration tracks:

```bash
bash ./scripts/migrate-casino-main.sh
bash ./scripts/migrate-casino.sh
```

Release / smoke checks for casino:
- `make health-casino`
- `go test -run Integration -v ./services/casino/internal/casino` with `MUDRO_CASINO_INTEGRATION_TEST_DSN` set to an isolated migrated casino DB
- wallet-sync scenarios via `docs/casino-smoke-checklist.md`

## Frontend

Frontend lives in [`frontend/`](frontend/).

```bash
cd frontend
npm.cmd install
npm.cmd run dev
```

Additional frontend commands:

```bash
npm.cmd run lint
npm.cmd run build
```

The frontend dev server proxies API requests to `http://127.0.0.1:8080`.

See also: [`frontend/README.md`](frontend/README.md)

## Import, Backfill, and Maintenance

The repository contains operational CLI tooling under [`tools/`](tools/), including:

- Telegram imports
- VK imports
- comment/media backfill
- dedupe and merge maintenance utilities

Examples:

```bash
make tg-csv-import CSV=/path/to/export.csv
make tg-comments-csv-import CSV=/path/to/comments.csv
make tg-comment-media-import DIR=/path/to/media
make comment-backfill
make media-backfill
```

## Direct Entrypoints

When you need to run services directly:

```bash
go run ./services/feed-api/cmd
go run ./services/agent/cmd
go run ./services/bot/cmd
go run ./services/casino/cmd/casino
```

Useful make targets:

```bash
make bot-run
make agent-plan-once
make agent-plan
make agent-work
make casino-run
```

## Legacy and Transitional Areas

The repository still contains compatibility paths that are intentionally not the primary release entrypoints:

- [`legacy/old/`](legacy/old/)
- [`cmd/`](cmd/)
- [`ops/compose/docker-compose.legacy.yml`](ops/compose/docker-compose.legacy.yml)

Treat them as recovery or compatibility paths unless a task explicitly targets them.

## Release References

- [`CHANGELOG.md`](CHANGELOG.md)
- [`docs/release-showcase-checklist.md`](docs/release-showcase-checklist.md)
- [`docs/casino-smoke-checklist.md`](docs/casino-smoke-checklist.md)
- [`docs/casino-db-stabilization-checklist.md`](docs/casino-db-stabilization-checklist.md)
- [`ops/runbooks/ops-runbook.md`](ops/runbooks/ops-runbook.md)
- [`Makefile`](Makefile)
