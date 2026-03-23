# Agent Workflows

This document captures the practical multi-agent workflow for MUDRO after adding `tools/opus-gateway`.

## Recommended Split

Use the available agents intentionally instead of asking every model to do every job.

- `gpt-5.4`
  Best for architecture review, repo exploration, risk analysis, and independent audits.
- `gpt-5.3-codex`
  Best for narrow-scope code changes, iterative fixups, and local file edits in a bounded area.
- `opus-gateway`
  Use when you explicitly want Anthropic `Opus` with your own local `ANTHROPIC_API_KEY` as a sidecar HTTP service.

## Practical Loop

1. Use `gpt-5.4` for repo exploration or architectural review.
2. Hand a narrow, file-bounded change set to `gpt-5.3-codex`.
3. Use `opus-gateway` only when a second external model is materially useful for comparison, code review, or long-form implementation help.
4. Keep each worker on a disjoint file set when possible.
5. Merge results back into the main branch only after local review.

## When To Use Opus Gateway

Use `opus-gateway` for:
- second-opinion architecture reviews
- independent implementation suggestions
- local coding help that must run on your own Anthropic key

Do not treat it as a native replacement for the built-in subagents of this Codex chat. It is a separate local sidecar process with its own HTTP interface.

## Safety Notes

- Keep `ANTHROPIC_API_KEY` outside the repository.
- Run `opus-gateway` only on `127.0.0.1`.
- Prefer `read-only` mode first and enable `edit` only for bounded tasks.
- Enable `allowBash` only when the task truly benefits from the small allowlist.
