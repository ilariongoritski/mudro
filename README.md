# MUDRO

MUDRO is a microservices-first monorepo for feed ingestion, agent workflows, Telegram operations, casino runtime, frontend UI, and data import tooling.

The repository is already in an active transition to `services/*`, `tools/*`, and `ops/*`. The canonical local runtime is the `core` contour under `ops/compose`, while legacy and transitional paths are still present for compatibility and recovery.

## Canonical State

Active runtime services:
- `services/feed-api`
- `services/agent`
- `services/bot`
- `services/casino`

Additive and migration-stage services:
- `services/api-gateway`
- `services/bff-web`
- `services/auth-api`
- `services/orchestration-api`

Key repository zones:
- `services/` long-running services
- `tools/` import, backfill, and maintenance CLI
- `internal/` shared Go packages
- `contracts/` HTTP and event contracts
- `migrations/` SQL migrations
- `frontend/` React + TypeScript + Vite application
- `ops/` compose, runbooks, scripts, nginx, systemd
- `legacy/old/` old runtime pieces kept for transition and reference

## Requirements

- Docker and Docker Compose
- Go `1.24`
- Node.js and npm for `frontend/` and `tools/opus-gateway`
- PostgreSQL is normally started through Docker Compose
- Windows + WSL2 is a supported local development setup

Primary local DSN:

```text
postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable
```

## Quick Start

Canonical local bootstrap:

```bash
make core-up
make dbcheck-core
make migrate-runtime
make tables-core
make health-runtime
```

Useful follow-up checks:

```bash
make core-ps
make test-active
make count-posts-core
```

Expected local endpoints:
- API health: `http://127.0.0.1:8080/healthz`
- Frontend dev server: `http://127.0.0.1:5173`

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

## Runtime Contours

Use the right contour for the right task:

- [`ops/compose/docker-compose.core.yml`](ops/compose/docker-compose.core.yml)
  Canonical local runtime for `db`, `redis`, `kafka/redpanda`, `api`, and `agent`.
- [`ops/compose/docker-compose.services.yml`](ops/compose/docker-compose.services.yml)
  Additive microservice layer for `auth-api`, `orchestration-api`, `bff-web`, and `api-gateway`.
- [`docker-compose.yml`](docker-compose.yml)
  Separate simplified contour for `db + casino-db + casino-api`. Useful for focused casino work and some legacy/recovery scenarios, but not the main entry path for the whole runtime.

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

## Opus Gateway

MUDRO includes a local HTTP sidecar for running `Opus` against this repository using your own `ANTHROPIC_API_KEY`.

Location:
- [`tools/opus-gateway/`](tools/opus-gateway/)

What it does:
- listens on `127.0.0.1:8788` by default
- uses the official `@anthropic-ai/claude-code` SDK
- defaults to `claude-opus-4-1-20250805` unless overridden by `OPUS_GATEWAY_MODEL` or `ANTHROPIC_MODEL`
- keeps the agent inside the repo root
- supports `read-only` and `edit` runs
- optionally enables a tightly allowlisted `Bash`

Install and run from the repo root:

```bash
npm run opus-gateway:install
npm run opus-gateway
```

Environment:

```text
ANTHROPIC_API_KEY=...
OPUS_GATEWAY_PORT=8788
OPUS_GATEWAY_MODEL=claude-opus-4-1-20250805
```

Health check:

```bash
curl http://127.0.0.1:8788/healthz
```

Run a task:

```bash
curl -X POST http://127.0.0.1:8788/v1/run ^
  -H "Content-Type: application/json" ^
  -d "{\"prompt\":\"Объясни структуру services/feed-api\",\"mode\":\"read-only\"}"
```

Important:
- this gateway is a local sidecar HTTP service, not a native extension point for the built-in subagents of this Codex chat
- `ANTHROPIC_API_KEY` must stay outside the repository
- `read-only` runs go through Claude Code `plan` permission mode; `edit` runs use `acceptEdits`
- logs are written under `var/log/opus-gateway/`

See also: [`tools/opus-gateway/README.md`](tools/opus-gateway/README.md)

## Legacy and Transitional Areas

This repository is still in migration. The following areas exist intentionally:

- [`legacy/old/`](legacy/old/)
- [`cmd/`](cmd/)
- [`ops/compose/docker-compose.legacy.yml`](ops/compose/docker-compose.legacy.yml)

Treat them as compatibility or recovery paths unless a task explicitly targets them.

## Where To Look Next

- [`docs/service-catalog.md`](docs/service-catalog.md)
- [`docs/repository-topology.md`](docs/repository-topology.md)
- [`docs/repo-layout.md`](docs/repo-layout.md)
- [`docs/agent-workflows.md`](docs/agent-workflows.md)
- [`ops/runbooks/ops-runbook.md`](ops/runbooks/ops-runbook.md)
- [`Makefile`](Makefile)
