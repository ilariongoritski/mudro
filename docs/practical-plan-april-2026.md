# Practical Plan: April 2026

## Objective
Finalize the MUDRO MVP codebase, stabilize the Casino microservice, and prepare for the Feed rollout.

## Phase 1: Stability (COMPLETED)
- [x] Fix Casino backend test regressions.
- [x] Normalize Vercel infrastructure (clean rewrites).
- [x] Secure sensitive env variables in `.env.example`.
- [x] Fix Plinko balance synchronization.

## Phase 2: Feed & Media (IN PROGRESS)
- [ ] Run full media backfill into MinIO.
- [ ] Populate feed with 50+ promotional items via `tgload`.
- [ ] Validate `/media` proxying across all environments.

## Phase 3: Casino Polish (PLANNED)
- [ ] Implement betting limits (Min/Max) in the UI.
- [ ] Finalize Roulette "Result display" sequence.
- [ ] End-to-end integration test for Roulette streaming.

## Operational Notes
- Use `make health-mvp` for a full system check.
- Keep `main` branch protected and verified via `test-active`.
