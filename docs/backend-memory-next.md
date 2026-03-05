# Backend Memory (What Is Missing for Front)

Updated: 2026-03-05

## Current Gap
Frontend is ready, but backend data context is incomplete in the active DB container.

## Facts Checked
- API endpoint is up: `GET /healthz` returns `{"status":"ok"}`.
- Current DB used by API (`localhost:5433` -> `mudro11-db-1`) has:
  - `total_posts = 0`
  - `posts_with_media = 0`
- Therefore `/api/front` returns empty feed.

## What Backend Needs Next
1. Restore/import post data into the active DB (`mudro11-db-1`), including `media` JSON.
2. Confirm migrations and data integrity:
   - `posts`
   - `post_reactions`
   - `post_comments`
3. Verify API responses:
   - `/api/front?limit=10` non-empty
   - `/api/posts?limit=10&page=1` non-empty
4. Ensure media URLs are reachable for browser render (not local-only paths).

## Environment Gaps to Remove
- Install local tooling in host shell:
  - `make`
  - `go`
  - `psql`
- This will allow full native health-loop without docker command workarounds.
