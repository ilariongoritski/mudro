# Backend Memory (Current MVP Gap)

Updated: 2026-03-22

## Current Status
The canonical runtime is up on the main stack, including the feed API and the casino subsystem. The old `mudro11-db-1` reference is no longer the canonical source of truth for current MVP checks.

## Facts Checked
- API health is OK: `GET /healthz` returns `{"status":"ok"}`.
- Canonical main DB is exposed on `localhost:5433`.
- Casino DB is exposed on `localhost:5434`.
- Casino API health is OK on `localhost:8082`.
- Frontend preview/build is working against the current main stack.

## What Still Needs Attention for MVP
1. Keep the demo seed/backfill path working on fresh installs so the feed is not empty after a clean bootstrap.
2. Keep media URLs reachable through localhost/VPS/Vercel proxy routes.
3. Keep docs and runbooks aligned with the current canonical ports:
   - main DB: `5433`
   - casino DB: `5434`
   - main API: `8080`
   - casino API: `8082`
4. Avoid using stale container names like `mudro11-db-1` in canonical docs unless explicitly discussing a legacy local test environment.

## Environment Gaps to Remove
- For native checks, it is still useful to have host tooling available:
  - `make`
  - `go`
  - `psql`
- That keeps the health loop reproducible without ad-hoc Docker shell workarounds.
