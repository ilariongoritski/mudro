# ADR 0004: Shared Postgres Runtime

Date: 2026-06-30

## Status

Accepted

## Context

Mudro now has several Go services around one product runtime:

- `feed-api` is the main MVP HTTP API.
- `casino` is a separate service with its own runtime boundary.
- `movie-catalog` is an additive catalog service and import/migration contour.

Splitting every service into a separate physical Postgres instance would add local setup cost, CI complexity, and rollout friction before the product needs that isolation.

## Decision

`feed-api`, `casino`, and `movie-catalog` use one Postgres installation/runtime contour by default.

Services can still keep separate tables, schemas, migrations, DSN variables, and rollout scripts. This keeps ownership boundaries explicit without multiplying local infrastructure.

The canonical local/runtime migration inventory is:

- main app migrations from `migrations/*.sql`, excluding `*.down.sql`;
- movie-catalog migrations from `migrations/movie_catalog/*.sql`;
- casino migrations from `services/casino/migrations/*.sql`, excluding `*.down.sql`.

## Consequences

- Local development uses one database stack first, with service-specific migrations applied explicitly.
- CI must validate migration inventories and service tests without scanning unrelated directories such as `node_modules`.
- Data-destructive schema cleanup, including legacy JSONB media removal, must be guarded and run through up-only migrations on isolated/staging databases before production.
- A future move to separate physical databases remains possible, but it is a deployment decision, not the default architecture now.
