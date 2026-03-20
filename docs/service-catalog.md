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

## Old / Legacy

1. `reporter-old`
- Path: `legacy/old/services/reporter-old/cmd`
- Run: `go run ./legacy/old/services/reporter-old/cmd`
- Status: **Old (not used by default)**.
- Enable only through `ops/compose/docker-compose.legacy.yml` or `make legacy-report-run`.
