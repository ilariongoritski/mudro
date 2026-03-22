# Development Plan

## Overview
This plan maps Skaro workflow onto the real `mudro` roadmap. The goal is not a greenfield rewrite, but controlled progress on data correctness, runtime safety, and the external MVP surface.

## 01 — Local Workflow and Safety
_Directory: `milestones/01-local-workflow-and-safety/`_

| # | Task | Status | Dependencies | Description |
|---|------|--------|--------------|-------------|
| 1 | skaro-bootstrap-local | done | — | Initialize `.skaro/` for `mudro`, wire roles, verify commands, constitution, architecture, and plan. |
| 2 | stabilize-go-verify | planned | skaro-bootstrap-local | Make repo-level Go verification deterministic for Skaro by excluding or fixing scratch-only blockers like `tmp/feed_check.go`. |
| 3 | codify-local-runbook | planned | skaro-bootstrap-local | Add Skaro-facing notes for local run flow, shell expectations, and secret handling. |

## 02 — Data Correctness and API Reliability
_Directory: `milestones/02-data-correctness/`_

| # | Task | Status | Dependencies | Description |
|---|------|--------|--------------|-------------|
| 1 | finish-tg-canonicalization | planned | stabilize-go-verify | Reduce remaining ambiguous Telegram discussion-root leftovers safely, without false merges. |
| 2 | lock-api-data-invariants | planned | finish-tg-canonicalization | Convert known comment/media/API invariants into repeatable checks and docs. |
| 3 | review-import-contracts | planned | lock-api-data-invariants | Re-check CSV/JSON importer expectations and document supported source formats. |

## 03 — Product MVP and VPS Runtime
_Directory: `milestones/03-product-mvp/`_

| # | Task | Status | Dependencies | Description |
|---|------|--------|--------------|-------------|
| 1 | vps-https-front | planned | lock-api-data-invariants | Put the self-hosted VPS frontend behind a normal HTTPS reverse proxy and re-evaluate external `:8080`. |
| 2 | frontend-mvp-polish | planned | vps-https-front | Finish remaining mobile/detail-drawer/empty-state polish for the public feed. |
| 3 | reporter-agent-hardening | planned | frontend-mvp-polish | Tighten runtime monitoring and operational recovery paths for agent/reporter/bot contours. |

## Backlog

| # | Task | Status | Description |
|---|------|--------|-------------|
| 1 | adr-backfill-key-decisions | idea | Backfill ADRs for the most important already-made decisions: media normalization, comment model, VK snapshot-only, VPS self-hosted frontend. |
| 2 | local-llm-fallback | idea | Revisit whether a local fallback LLM contour is worth enabling for reviewer-only or emergency scenarios. |
| 3 | public-domain-and-tls | idea | Finalize domain/TLS layer for the external MVP when domain is available. |

---

## Status Legend
- **idea** — not yet scoped
- **planned** — scoped, assigned to milestone
- **in-progress** — actively being developed
- **done** — completed and reviewed
- **cut** — removed from scope (with reason)

## Change Log
- 2026-03-17: Initial `mudro` Skaro plan created from existing repo state and `.codex` priorities.
