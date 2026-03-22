# Imported Snapshot From `.codex`

**Imported at:** 2026-03-17  
**Source of truth:** `.codex/todo.md`, `.codex/done.md`, `.codex/top10.md`, `.codex/state.md`

## Current project shape
- `mudro` already has a working content platform contour: Go backend, PostgreSQL, Telegram/VK import paths, React frontend, bots, and an agent/reporter process layer.
- `.codex` remains the detailed operational memory and must not be replaced by Skaro.
- The current automation worktree is responsible for process, workflow, templates, task specs, repo hygiene, and local developer flow.

## Durable decisions already made
- Telegram is the live source; VK is snapshot-only.
- Media is moving through a normalized-first model.
- Comment graph and comment reactions are now treated as first-class data structures.
- The public MVP can be served directly from VPS via `nginx`; Vercel is no longer the only external surface.

## Active process priorities
- Reduce chaos in local workflow and documentation.
- Make repeated starts and worktree-based parallel work predictable.
- Keep reviewable summaries in Skaro while preserving raw history in `.codex`.
- Turn implicit decisions into explicit ADRs and task specs.

## Boundaries for this worktree
- Routine product runtime bugs belong to `mudro11-bugs`.
- VPS/deploy/runtime operations belong to `mudro11-devops`.
- This worktree should improve tooling, process, templates, and task framing.
