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

## Transitional / Compatibility

- `cmd/` — temporary forwarding stubs to `tools/*`.

## Archived

- `legacy/old/` — non-default runtime and deprecated assets.
  - `legacy/old/cmd-runtime/*`
  - `legacy/old/services/reporter-old/*`
  - `legacy/old/misc/*`
  - `legacy/old/manifest.yaml`
