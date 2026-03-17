# Skaro Docs

This directory stores stable summaries and imported context for the `mudro` cockpit.

## Rules
- Keep raw operational history in `.codex/*`.
- Import only condensed summaries into `.skaro/docs/imported/`.
- Keep durable user guides and process notes here, not screenshots, logs, or dumps.
- Do not store secrets, tokens, or raw environment snapshots here.

## Suggested split
- `imported/` - snapshots distilled from `.codex` memory and other stable sources
- top-level docs - operator notes, cockpit usage, and process-oriented guides

## Current intent
This worktree uses Skaro for:
- constitution
- architecture
- milestones/tasks/specs
- ADRs
- process and review guidance

This worktree does not use Skaro as a replacement for:
- `.codex` memory
- raw logs
- runtime deployment output
