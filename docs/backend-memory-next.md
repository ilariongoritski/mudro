# Backend Memory (Current MVP Gap)

Updated: 2026-04-17

## Current Status
- **Canonical Runtime**: Stable on main stack. Feed API and Casino subsystems are operational.
- **Fairness Engine**: Unified HMAC-SHA512 engine implemented for Casino (Roulette, Plinko, Blackjack).
- **Infrastructure**: Vercel rewrites normalized (broad wildcard removed). Docker stack uses `ops/compose/`.
- **Testing**: `make test-active` now runs unit tests and contract validation. Casino `handlers_test.go` fixed.

## Facts Checked
- API health OK: `GET /healthz` returns `{"status":"ok"}`.
- Casino API health OK on `localhost:8082`.
- Casino tests passing with context propagation.
- Environment variables for Casino constraints (`CASINO_RTP_BPS`, `CASINO_MAX_BET`) added.

## What Still Needs Attention for MVP
1. **Feed MVP**: Finalize media backfill and prompt-driven feed population.
2. **Media Reachability**: Ensure `/media` proxy in Vercel and local dev is consistent with MinIO.
3. **Casino Polish**: Add betting limits enforcement in the UI (synced with backend constraints).
