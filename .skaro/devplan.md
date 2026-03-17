# Development Plan

## Overview
This Skaro plan treats `.skaro/` as the project cockpit for `mudro`, while `.codex/` stays the detailed operational memory. The immediate goal in this worktree is to reduce process chaos, make local execution reproducible, and establish task-spec and ADR discipline without mixing in routine product bugfix work.

## 01 - Project Cockpit
_Directory: `milestones/01-project-cockpit/`_

| # | Task | Status | Dependencies | Description |
|---|------|--------|--------------|-------------|
| 1 | skaro-workflow-hardening | in-progress | - | Align `config.yaml`, VS Code tasks, port conventions, and secret handling so Skaro is usable every day from this worktree. |
| 2 | codex-skaro-memory-sync | planned | skaro-workflow-hardening | Import stable summaries from `.codex` into `.skaro/docs/imported/` without replacing the source memory system. |
| 3 | adr-backfill-key-decisions | planned | codex-skaro-memory-sync | Backfill ADRs for already-made durable decisions so they stop living only in chat and ad hoc notes. |

## 02 - Public MVP Release
_Directory: `milestones/02-public-mvp-release/`_

| # | Task | Status | Dependencies | Description |
|---|------|--------|--------------|-------------|
| 1 | public-mvp-surface-map | planned | adr-backfill-key-decisions | Map the externally visible MVP surfaces, their owners, and the canonical review links for product acceptance. |
| 2 | release-review-checklists | planned | public-mvp-surface-map | Define lightweight pre-release and external-review checklists for feed, API, data freshness, and documentation. |

## 03 - VPS Runtime
_Directory: `milestones/03-vps-runtime/`_

| # | Task | Status | Dependencies | Description |
|---|------|--------|--------------|-------------|
| 1 | runtime-runbook-convergence | planned | adr-backfill-key-decisions | Converge `.codex`, Skaro docs, and runbooks on one operational story for local recovery, VPS runtime, and boundaries. |
| 2 | agent-ops-observability | planned | runtime-runbook-convergence | Capture the agent/reporter/bot operational contour, review gates, and observability expectations in stable docs and task specs. |
| 3 | secure-skaro-hosting-policy | planned | runtime-runbook-convergence | Record the rule that any future VPS-hosted Skaro UI requires access control and must never expose secrets or raw memory publicly. |

## Backlog

| # | Task | Status | Description |
|---|------|--------|-------------|
| 1 | worktree-parallel-flow | idea | Document how to run Skaro cleanly across multiple worktrees without path confusion or accidental shared-state assumptions. |
| 2 | imported-doc-refresh | idea | Add a lightweight refresh cadence for `.skaro/docs/imported/` snapshots from `.codex`. |
| 3 | skaro-ci-validation | idea | Decide whether Skaro validation should run in CI for selected task types and what should remain local-only. |

---

## Status Legend
- **idea** - not yet scoped
- **planned** - scoped and waiting
- **in-progress** - currently being shaped in this worktree
- **done** - completed and reviewed
- **cut** - intentionally removed from scope

## Change Log
- 2026-03-17: Reshaped the Skaro plan around project cockpit, public MVP release surface, and VPS runtime process documentation.
