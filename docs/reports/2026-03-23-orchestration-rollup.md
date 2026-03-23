# MUDRO Rollup: chat, subagents, and VPS rollout

Date: `2026-03-23`

## Scope

This report captures the concrete outcomes of the current Codex-controlled orchestration loop across:

- repository hardening and systemd tracking;
- VPS rollout of tracked MUDRO services;
- worker-plane rollout for `OpenClaw` and `Skaro`;
- reviewer/checklist passes from local subagents and Claude-backed role runs.

It intentionally omits secrets, runtime tokens, and raw environment values.

## Repository outcomes already landed before this report

The following tracked runtime milestones were completed and pushed to `main` during this thread:

- `a916100` `Track systemd runtime for mudro-api`
- `52f0622` `Track hardened systemd runtimes for bot, agent, OpenClaw, and Skaro`
- `370ee06` `Fix tracked OpenClaw systemd rollout`

These commits established:

- tracked `systemd` templates for `mudro-api`, `mudro-bot`, `mudro-agent-worker`, `mudro-agent-planner`, `mudro-agent-planner.timer`;
- tracked `systemd` templates for `openclaw.service` and `skaro.service`;
- tracked runtime env examples outside the repo-secret boundary;
- tracked installers for core MUDRO services and the worker plane;
- the repo-side correction needed after live VPS rollout of `OpenClaw`.

## VPS rollout summary

### Core MUDRO runtime

The tracked core stack was applied on VPS and verified live:

- `mudro-api.service`
- `mudro-bot.service`
- `mudro-agent-worker.service`
- `mudro-agent-planner.timer`

Live checks that passed during rollout:

- `systemctl is-active` returned `active` for all tracked core units;
- `http://127.0.0.1:8080/healthz` responded successfully;
- `http://127.0.0.1:8080/api/front?limit=1` returned feed JSON;
- external frontend requests to the Vercel-hosted UI also returned feed data successfully.

### Worker plane

The tracked worker plane was also applied on VPS and verified live:

- `openclaw.service`
- `skaro.service`

Live checks that passed during rollout:

- `http://127.0.0.1:4700/` returned the Skaro dashboard HTML;
- `http://127.0.0.1:18789/` returned `200 OK` for the OpenClaw gateway;
- `systemctl is-active` returned `active` for both units after the tracked rollout fix.

## Media and deployed UI state

During the same thread, the deployed frontend and media path were brought back to a working state:

- frontend API requests on the public Vercel site returned data;
- media GET requests returned image bodies;
- the runtime media layout was aligned with the database/media model already tracked in the application.

## Subagent and reviewer rollup

### Local Codex subagents

The following local subagent roles were used as bounded review/checklist helpers:

- `Franklin`
  - focus: env/runtime prerequisites for `mudro-api`, `mudro-bot`, `mudro-agent`
  - key useful signal: placeholder-sensitive envs and runtime prerequisites needed to be kept strictly outside the repository

- `Zeno`
  - focus: prerequisites for `OpenClaw` / `Skaro`
  - key useful signal: tracked VPS wrappers and service units were necessary, but the actual runtime env remained a server-side responsibility

These local subagents were treated as reviewers only. Final diffs and rollout decisions remained under Codex control.

### Claude-backed role passes

Claude-backed role runs were used as planning/review loops rather than direct code authorship. Relevant run groups across this thread included:

- `20260323-systemd-template`
- `20260323-systemd-worker-plane`
- `20260323-systemd-final-review`
- `20260323-vps-live-rollout`

Roles used across these runs included combinations of:

- `devops`
- `security`
- `integration`
- `architect`
- `frontend`
- `tester`

The useful outputs from these runs were narrowed to:

- hardening/service-layout checklists;
- warnings about env/secrets boundaries;
- rollout validation order and smoke checks.

No Claude output was applied blindly. Repo changes were accepted only after local validation and live VPS verification.

## Decisions taken in this thread

1. `Codex` remained the control plane for repository truth and rollout decisions.
2. `OpenClaw` / `Skaro` were treated as a separate worker plane rather than being mixed into the core MUDRO runtime.
3. Runtime secrets were kept out of tracked files; tracked files describe contract and install flow only.
4. Live VPS drift discovered during rollout was pulled back into tracked repository state with follow-up fixes.

## Remaining gap after this report

At the time of writing this report, one operational gap remained to be closed in tracked form:

- a tracked `claudeusageproxy.service` for VPS-side token accounting used by `Skaro` / `OpenClaw`

That gap is intentionally addressed in a separate change so the reporting commit stays readable and reversible.

## Notes on repository hygiene

- This report does not include raw logs, `.env` data, or runtime dumps.
- This report is intended to be safe to keep in `main`.
