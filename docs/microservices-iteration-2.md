# Microservices Iteration 2

This document describes the next safe landing after the first bootstrap step.

## What lands in iteration 2

1. `services/auth-api`
   - thin extraction of authentication and admin HTTP routes
   - keeps using the existing `internal/auth` domain service and the current Postgres schema
   - exposes:
     - `/api/v1/auth/*`
     - `/api/v1/admin/*`

2. `services/orchestration-api`
   - thin extraction of orchestration status as a dedicated service boundary
   - keeps the current logic additive and backward-compatible
   - exposes:
     - `/api/v1/orchestration/status`

3. `services/api-gateway`
   - now routes:
     - `/api/v1/auth/*` -> `auth-api`
     - `/api/v1/admin/*` -> `auth-api`
     - `/api/v1/orchestration/*` -> `orchestration-api`
     - all other `/api/v1/*` -> `feed-api`

## Why this is safe

- No schema change
- No DB split yet
- No frontend big-bang switch required
- Existing `feed-api` remains the source of truth for feed routes
- `auth-api` and `orchestration-api` are additive thin services around existing logic

## Validation target

Minimum validation for iteration 2:

```bash
go test ./services/api-gateway/... ./services/auth-api/... ./services/orchestration-api/... ./tools/validate-contracts
go run ./tools/validate-contracts -dir ./contracts
docker compose -f ops/compose/docker-compose.core.yml -f ops/compose/docker-compose.services.yml config
python -m py_compile scripts/claude/run_role_matrix.py
```

## What does not happen yet

- no auth domain rewrite
- no orchestration state rewrite
- no database split
- no removal of legacy `/api/auth/*` or `/api/orchestration/status`
- no direct frontend cutover to the gateway as the only runtime path
