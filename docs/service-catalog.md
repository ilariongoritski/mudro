# Service Catalog

## Active Services

1. `feed-api`
- Path: `services/feed-api/cmd`
- Run: `go run ./services/feed-api/cmd`
- Purpose: HTTP API, feed endpoints, health, media proxy.

2. `agent`
- Path: `services/agent/cmd`
- Run: `go run ./services/agent/cmd --mode <planner|worker|once|approve|reject>`
- Purpose: planner/worker automation and task lifecycle events.

3. `bot`
- Path: `services/bot/cmd`
- Run: `go run ./services/bot/cmd`
- Purpose: Telegram control-plane commands for project operations.

4. `casino`
- Path: `services/casino/cmd`
- Run: `go run ./services/casino/cmd`
- Purpose: isolated casino runtime with its own Postgres database, RTP/admin settings, and separate API surface.

## Bootstrap / Migration Services

1. `api-gateway`
- Path: `services/api-gateway/cmd`
- Run: `go run ./services/api-gateway/cmd`
- Purpose: additive public edge skeleton for `/api/v1/*` and BFF routing without breaking the current `feed-api` runtime.

2. `bff-web`
- Path: `services/bff-web/cmd`
- Run: `go run ./services/bff-web/cmd`
- Purpose: additive web aggregation layer for `/api/bff/web/v1/*`, starting with timeline, orchestration status, and casino widget aggregation.

3. `auth-api`
- Path: `services/auth-api/cmd`
- Run: `go run ./services/auth-api/cmd`
- Purpose: additive auth/admin service boundary that preserves the current JWT and user runtime while moving `/api/v1/auth/*` and `/api/v1/admin/*` out of the legacy feed entrypoint.

4. `orchestration-api`
- Path: `services/orchestration-api/cmd`
- Run: `go run ./services/orchestration-api/cmd`
- Purpose: additive orchestration-status boundary for `/api/v1/orchestration/*`, keeping the current status logic isolated behind a dedicated thin service.

## Old / Legacy

1. `reporter-old`
- Path: `legacy/old/services/reporter-old/cmd`
- Run: `go run ./legacy/old/services/reporter-old/cmd`
- Status: **Old (not used by default)**.
- Enable only through `ops/compose/docker-compose.legacy.yml` or `make legacy-report-run`.
