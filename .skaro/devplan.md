# Development Plan

## Overview
This plan turns Skaro into a useful project cockpit for `mudro` right now.
The goal is practical:
- preserve the existing `.codex` memory and logs
- expose the active backlog as milestones/tasks in Skaro
- keep one clear public MVP link for release
- prepare, but not yet force, a future VPS-hosted Skaro mode

## 01 - Project Cockpit
_Directory: `milestones/01-project-cockpit/`_

| # | Task | Status | Dependencies | Description |
|---|------|--------|--------------|-------------|
| 1 | import-project-memory | done | none | Import `.codex` memory snapshots and build Skaro-facing summaries. |
| 2 | stabilize-skaro-workflow | done | import-project-memory | Make local Skaro usage clear, commitable, and VS Code-friendly. |

## 02 - Public MVP Release
_Directory: `milestones/02-public-mvp-release/`_

| # | Task | Status | Dependencies | Description |
|---|------|--------|--------------|-------------|
| 1 | lock-public-mvp-links | done | none | Record the working VPS and Vercel public URLs and mark protected previews as non-final. |
| 2 | finalize-release-checklist | planned | lock-public-mvp-links | Keep a short final release checklist for the MVP handoff. |

## 03 - VPS Runtime
_Directory: `milestones/03-vps-runtime/`_

| # | Task | Status | Dependencies | Description |
|---|------|--------|--------------|-------------|
| 1 | stabilize-vps-runtime | planned | none | Document how Skaro fits the current VPS runtime as an internal cockpit. |
| 2 | prepare-skaro-vps-host | planned | stabilize-vps-runtime | Define a safe future VPS-hosted Skaro shape with internal-only access. |

## Backlog

| # | Task | Status | Description |
|---|------|--------|-------------|
| 1 | adr-backfill-key-decisions | planned | Backfill ADRs for already-made durable decisions: media normalization, comment model, VPS self-hosted frontend, VK snapshot-only. |
| 2 | public-domain-and-tls | planned | Finalize the domain and TLS layer for the public MVP when a stable domain is available. |
| 3 | skaro-vps-systemd-rollout | idea | Turn the future VPS-hosted Skaro cockpit into a managed `systemd` service. |

---

## Status Legend
- **idea** - not yet scoped
- **planned** - scoped and ready
- **in-progress** - actively being worked on
- **done** - completed in the current baseline

## Change Log
- 2026-03-17: Reworked the plan around the real `mudro` state: imported project memory, documented public MVP links, and created milestone/task structure for Skaro.
