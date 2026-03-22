# Claude Role Matrix for MUDRO

This repository already uses a local Claude-compatible proxy and OpenClaw/Skaro-oriented orchestration.

This document defines a reproducible role matrix for parallel extended-thinking reviews:

- `architect`
- `frontend`
- `backend`
- `tester`
- `devops`
- `security`
- `integration`
- `data`

## Purpose

Use the role matrix when the task is large enough to benefit from multiple specialist passes before Codex applies changes locally.

The role matrix does not write tracked files on its own.
It writes outputs under the local accounting/runtime root outside the repository by default:

- `D:\mudr\_mudro-local\claude-orch\runs\<run-id>\summary.json`
- `D:\mudr\_mudro-local\claude-orch\runs\<run-id>\<role>.md`

The proxy keeps usage accounting separately:

- `D:\mudr\_mudro-local\claude-orch\ledger\usage_log.jsonl`
- `D:\mudr\_mudro-local\claude-orch\ledger\token_usage.yaml`
- `D:\mudr\_mudro-local\claude-orch\ledger\role_usage.yaml`

## Run

```bash
python scripts/claude/run_role_matrix.py \
  --task "Review the next safe microservice iteration for MUDRO and list concrete changes." \
  --repo-root D:\mudr\mudro11
```

## Rules

1. All role prompts are kept in `ops/claude-workers/roles/`.
2. Keep prompts grounded in the actual Go repository.
3. Do not let role prompts assume Node/Express/Prisma unless that stack is explicitly present.
4. Always include the repository context pack from the checked-in service catalog, service map, and microservice iteration docs.
5. Use the role matrix for planning and review, not as a substitute for local validation.
6. Codex remains the control plane and the only actor that applies tracked diffs.
