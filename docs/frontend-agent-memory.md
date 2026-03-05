# Frontend Agent Memory

Updated: 2026-03-05  
Branch: `codex/frontend-mudro11-fsd`

## Delivered
- New frontend app in `frontend/` with stack:
  - React + TypeScript
  - Redux Toolkit
  - RTK Query
  - Vite
  - FSD structure (`app/pages/widgets/features/entities/shared`)
- Draft visual concept page: `docs/frontend-preview.html`
- Hotpink-centered UI theme with responsive cards.
- Feed features:
  - source filter (`all`, `vk`, `tg`)
  - sort (`desc`, `asc`)
  - limit selector
  - load from `/api/front`
  - pagination via `/api/posts?page=...`

## Runtime Notes
- This shell lacks `make`, `go`, `psql` on PATH.
- `npm` must be invoked as `npm.cmd` (PowerShell policy).
- API currently runs via Docker container (`mudro-api-local`) when needed.

## Next Frontend Steps
- Add router + dedicated post/details views.
- Add tests (unit + integration).
- Optionally wire `frontend/dist` into backend static serving.
