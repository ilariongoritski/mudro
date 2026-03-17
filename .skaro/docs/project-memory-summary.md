# Project Memory Summary

## Purpose
This file is the Skaro-facing summary of the existing `mudro` project memory.
It does not replace `.codex/*`. It gives Skaro a compact, repo-native snapshot
of the project's current state, priorities, and public links.

## Imported sources
Detailed source snapshots were copied into `.skaro/docs/imported/`:

- `codex-todo.md`
- `codex-done.md`
- `codex-top10.md`
- `codex-memory.json`
- `codex-time-runtime.json`
- `codex-tg-control.jsonl`
- `codex-log-directories.txt`

## Current product baseline
- Product: unified VK + Telegram feed with Go backend, PostgreSQL, React frontend.
- Local process memory lives in `.codex/*`.
- Public MVP exists in two forms:
  - self-hosted VPS frontend on `http://91.218.113.247/`
  - Vercel public MVP on `https://frontend-psi-ten-33.vercel.app`
- Newer auth-protected preview exists on Vercel, but it is not suitable as the
  final public MVP URL because it returns `401 Unauthorized`.

## High-value project decisions already made
- VK is snapshot-only; active import focus is Telegram.
- Media moved to a normalized model with fallback to legacy JSONB.
- Comment model now has parent links and normalized reactions.
- VPS frontend is self-hosted via `nginx`, while Vercel remains a useful public
  review surface.
- Postgres on VPS is hardened behind loopback and dedicated app credentials.

## Current practical priorities
1. Keep Skaro as the project cockpit, not as a runtime replacement for `mudro`.
2. Convert active `.codex` backlog into milestones/tasks that show up in Skaro UI.
3. Keep a stable public MVP URL for external review.
4. Prepare a safe VPS-hosted Skaro mode later, but keep local workflow primary.

## Public MVP links
- VPS: `http://91.218.113.247/`
- Vercel public MVP: `https://frontend-psi-ten-33.vercel.app`
- Vercel protected preview: `https://frontend-nv1pu0992-goritskimihail-2652s-projects.vercel.app`

## Notes
- Raw operational history remains in `.codex/*`.
- Skaro should read summaries, tasks, plans, and ADRs, not massive raw logs.

