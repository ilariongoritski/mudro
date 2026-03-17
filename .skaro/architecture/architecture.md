# Architecture

## Overview
`MUDRO` is a content platform built around a unified feed of Telegram and VK source data, with a Go backend, PostgreSQL storage, import pipelines, a React frontend, and operational bots. The current runtime is still close to a modular monolith, but the codebase already separates API, importers, bot/reporter flows, and agent-like maintenance contours.

## Components
- `cmd/api`
  - Serves JSON endpoints such as `/api/front`, `/api/posts`, `/healthz` and HTML feed surfaces.
- Importers
  - `cmd/vkimport`, `cmd/tgimport`, `cmd/tgload`, `cmd/tgcsvimport`, `cmd/tgcommentsimport`, `cmd/tgcommentscsvimport`, `cmd/tgcommentmediaimport`, `cmd/tgdedupe`, `cmd/tgrootmerge`.
  - Responsible for safe normalization and loading of source data.
- Bots and reporting
  - `cmd/bot` for Telegram control and project interaction.
  - `cmd/reporter` for digest/report flows.
- Agent contour
  - `cmd/agent` plus `internal/agent` for planner/worker patterns and review-gated execution.
- Frontend
  - `frontend/` React + TypeScript application consuming the live API.
- Ops and deployment
  - `Makefile`, `docs/ops-runbook.md`, `scripts/ops/*`, `docker-compose*.yml`.

## Data Storage
- Main relational storage: PostgreSQL.
- Core tables:
  - `posts`
  - `post_comments`
  - `post_reactions`
  - `comment_reactions`
  - `media_assets`
  - `post_media_links`
  - `comment_media_links`
  - `agent_queue`
- Data model strategy:
  - normalized-first media and comment graph
  - legacy JSON payloads tolerated only as compatibility/fallback layer where still needed
- Local canonical DSN:
  - `postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable`

## Communication
- Primary app path: HTTP JSON from frontend to `cmd/api`.
- Local operational path: `Makefile` + WSL shell + `docker compose`.
- Bots use Telegram APIs and internal project state.
- Event-driven split is planned; Kafka exists as an optional future/backbone direction, not a mandatory runtime dependency today.

## Infrastructure
- Local development:
  - repository on `D:\mudr\mudro11`
  - WSL path `/mnt/d/mudr/mudro11`
  - Docker Compose for DB
  - backend commands from WSL
  - frontend commands from Windows shell or WSL
- VPS:
  - `nginx` serves frontend and reverse-proxies `/api`, `/media`, `/healthz`
  - backend services run under systemd / local process management
  - database access should stay loopback-only and non-public

## External Integrations
- Telegram Bot API
- OpenAI API for LLM-backed features and automation
- Optional/public hosting layers such as Vercel are auxiliary, not canonical runtime anymore
- Optional local/backup LLM runtimes may exist later, but are not assumed in the default local setup

## Security
- No destructive DB operations without explicit approval.
- No secret material in tracked repo artifacts.
- Public API changes must be deliberate and reviewed.
- Local tests and scripts must not silently target production or shared DBs.
- Import and dedupe flows must preserve referential consistency for comments, media, and reactions.

## Known Trade-offs
- The project spans Windows and WSL, so operational commands must be explicit about which shell owns them.
- `tmp/` may contain scratch helpers and should not define canonical verification status.
- The system is not yet a fully split microservice platform; some docs describe future structure, while runtime remains intentionally pragmatic.
- `VK` is preserved as snapshot-only content, while active freshness work is focused on Telegram.
