# LangGraph observe-only preparation

## Purpose

This directory is a design and test boundary for the future LangGraph orchestrator.
It is intentionally **not a running service** and contains no production credentials,
Docker socket access, deployment actions, database migrations, or write-capable tools.

Hermes remains the Telegram gateway, owner-approval interface, skills and MCP runtime.
LangGraph will become the durable workflow engine only after this observe-only phase is
validated.

## Phase A scope

Allowed inputs:

- Docker Compose service status (read-only)
- internal `/healthz` HTTP status (read-only)
- repository `git status` and latest commit (read-only)
- backup verifier output (read-only)
- disk usage (read-only)

Allowed outputs:

- structured report
- a persisted workflow state through a future checkpointer
- owner-facing recommendation

Forbidden actions:

- deploy, restart, rollback, migrations, database writes
- Git commit, push, merge, branch mutation
- package installation, system configuration, Docker pruning
- reads of `.env`, credentials, auth files, or production data
- external network calls except owner-approved documentation lookup

## Proposed graph

```text
START -> collect_observations -> classify_health -> build_report -> END
```

The graph will be invoked with a unique `run_id` and `thread_id`. It must be
idempotent: repeating the same run produces a report only and never changes system state.

## Persistence plan

1. Start with local test-only SQLite state outside the production database.
2. Validate interruption/resume and duplicate-run behavior.
3. Before production use, create a dedicated Postgres role and a dedicated schema;
   no access to casino, auth, or application schemas.
4. Store approval events separately and require an unexpired owner approval for any
   future write/deploy node.

## Promotion gates

- Unit tests cover all branches and report schema.
- Mocked command runner proves no command contains mutating operations.
- A manual observe-only run emits an auditable report.
- Owner approves database/checkpointer credentials and the execution boundary.
- Only then add a scheduled observe-only worker.
