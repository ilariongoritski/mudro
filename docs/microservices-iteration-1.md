# Microservices Iteration 1

This document describes the first safe landing step from the current mixed runtime toward the target MUDRO microservice layout.

## What lands in iteration 1

1. `services/api-gateway`
   - public HTTP entry skeleton
   - stable `/api/v1/*` prefix
   - forwards current `feed-api` routes without changing the existing runtime

2. `services/bff-web`
   - web-focused aggregation layer
   - exposes `/api/bff/web/v1/*`
   - timeline endpoint reads from the current database through the existing Go domain service
   - orchestration and casino widget endpoints aggregate/proxy current runtime responses

3. Optional compose layer
   - `ops/compose/docker-compose.services.yml`
   - can be combined with `ops/compose/docker-compose.core.yml`

4. Parallel Claude role matrix
   - `scripts/claude/run_role_matrix.py`
   - `ops/claude-workers/roles/*.md`

## Why this is safe

- No schema change
- No frontend switch required
- No removal of legacy routes
- Existing `feed-api` and `casino` remain the source of runtime truth
- New services are additive and can be rolled back independently

## Validation target

Minimum validation for iteration 1:

```bash
go test ./services/api-gateway/... ./services/bff-web/... ./tools/validate-contracts
go run ./tools/validate-contracts -dir ./contracts
python -m py_compile scripts/claude/run_role_matrix.py
```

Optional compose validation:

```bash
docker compose -f ops/compose/docker-compose.core.yml -f ops/compose/docker-compose.services.yml config
```

## What does not happen yet

- no direct frontend migration to BFF
- no identity extraction
- no DB split for feed/query-write services
- no removal of compat routes
- no OpenClaw-only deployment cutover

Follow-up safe step: `docs/microservices-iteration-2.md`
