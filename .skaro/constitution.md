# Constitution: MUDRO

## Stack
- Backend: Go `1.22+` in repo, currently developed and tested from WSL toolchain.
- Frontend: React `19` + TypeScript `5` + Vite.
- Database: PostgreSQL `16` in Docker Compose on local port `5433`.
- Infrastructure: local `Windows + WSL + Docker Desktop`, server `Ubuntu VPS + systemd + nginx + Docker Compose`.

## Coding Standards
- Prefer simple, reliable solutions and minimal diffs.
- Do not rewrite working subsystems without explicit reason.
- Keep backend code idiomatic Go: explicit data flow, predictable SQL, small cohesive functions.
- Keep frontend aligned with the existing `frontend/` structure and current FSD-style organization.
- Text files must stay `UTF-8` with `LF`; respect `.editorconfig` and `.gitattributes`.
- Never commit generated artifacts, dumps, secrets, or transient local outputs.

## Testing
- Required for backend changes: relevant `go test` in WSL; for repo-level checks, prefer the configured verify commands.
- Required for DB-sensitive changes: `make dbcheck` and `make tables`.
- Required for frontend changes: `npm.cmd --prefix frontend run build` and `npm.cmd --prefix frontend run lint`.
- Do not run destructive tests against a live or shared database.
- If a task changes runtime behavior, verify the real API path or browser flow, not only unit tests.

## Constraints
- `Makefile` is the canonical source for local operational commands.
- For `make`, `docker compose`, `go test`, migrations, and DB checks use the WSL/Linux contour.
- Risky DB actions require explicit human approval:
  - `drop`, `truncate`, reset, destructive migration rewrites, `docker compose down -v`.
- Public JSON API changes require explicit approval and clear acceptance criteria.
- Source policy:
  - `Telegram` is the live source.
  - `VK` is snapshot-only unless explicitly reopened.

## Security
- Never commit `.env`, `.skaro/secrets.yaml`, `data/`, `out/`, `var/log/`, or private dumps.
- Never print or copy API keys into tracked files.
- Prefer loopback-only DB exposure on VPS and dedicated app roles instead of superuser access.
- Preserve test safety: no implicit DSN that can touch a working DB by accident.

## LLM Rules
- Read `AGENTS.md` and the `.codex` memory files before meaningful work.
- Do not make hidden assumptions when data, migrations, API contracts, or infrastructure are involved.
- Ask before any change that may destroy data, reshape schema, or alter public behavior significantly.
- Do not overwrite or revert existing user changes unless explicitly requested.
- For medium and large tasks, create or update a task spec before implementation.
- For project-level decisions, record them as ADRs instead of burying them in chat only.
- If `AI_NOTES` are required by Skaro flow, keep them task-scoped and factual; do not use them as a substitute for `.codex` memory or ADRs.
