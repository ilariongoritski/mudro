# AGENTS — MUDRO

Canonical runbook + rules for coding agents.

## Commands (do not guess)
- deps up: `make up`
- deps down: `make down`
- tests: `make test`
- logs: `make logs`

## Rules
- Never push directly to `main` unless explicitly instructed.
- Prefer branch + PR workflow.
- Before any DB schema change/migration that may affect data: STOP and ask.
- Do not log secrets (DSN/password/tokens).
