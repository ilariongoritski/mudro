# Invariants

## Data Safety
- Do not run destructive DB actions without explicit approval.
- Do not introduce implicit DSN behavior that can target a working local or VPS database.
- Keep comment/media/reaction links referentially consistent during imports, merges, and dedupe passes.

## Source Policy
- Telegram is the live source of truth for active sync.
- VK is snapshot-only unless explicitly reopened.

## API Stability
- Do not break existing JSON API fields without explicit acceptance.
- Any compatibility bridge must be additive-first where feasible.

## Runtime Discipline
- Local DB checks and migrations use the WSL + Docker Compose contour.
- Frontend builds must stay reproducible from the checked-in `frontend/` project.
- Scratch files in `tmp/` are not product runtime and must not become canonical dependencies.

## Repo Hygiene
- Secrets remain outside tracked files.
- Generated outputs do not belong in commits.
- `.codex` remains the project memory system; Skaro artifacts complement it, they do not replace it.
