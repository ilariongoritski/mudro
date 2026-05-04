# Архитектурный обзор MUDRO к pre-MVP

Дата: 2026-05-03.

## Текущая форма системы

MUDRO сейчас является monorepo с Go-first backend, React/Vite frontend, SQL migrations, compose-based local/prod runtime и набором import/backfill CLI.

Фактический MVP runtime:

- `feed-api` — основной публичный API для ленты, auth/admin, chat и casino proxy.
- `casino` — отдельный сервис с собственной БД, внутренней авторизацией через `CASINO_INTERNAL_SECRET`.
- `frontend` — клиентский UI, локально ходит через Vite proxy в `feed-api`.
- `agent` и `bot` — операционный контур, не обязательный для первого показа ленты/casino, но часть целевой системы.

Production target: VPS/self-hosted через `docker-compose.prod.yml` + nginx/HTTPS.

## Что было структурно исправлено

- Casino HTTP слой разделён:
  - `http_router.go` — handler/router/health;
  - `http_middleware.go` — internal auth и user rate limiter;
  - `handlers.go` — только endpoint handlers.
- Casino store слой частично разделён:
  - `store_fairness.go` — server/client seed операции;
  - `store_blackjack.go` — blackjack state/start/action/resolve;
  - `store_errors.go` — Postgres error helpers;
  - `store.go` — core wallet/feed/roulette/plinko store.
- Добавлен отдельный local casino compose: `ops/compose/docker-compose.casino.local.yml`.
- Local core compose теперь прокидывает `CASINO_INTERNAL_SECRET` в `feed-api`, чтобы casino proxy не расходился с новым required-secret поведением.

## Крупные файлы и следующий разрез

Оставшиеся крупные файлы:

- `services/casino/internal/casino/store.go` — всё ещё основной кандидат на дальнейший разрез.
- `frontend/src/features/casino/api/casinoApi.ts` — стоит делить по endpoint groups: wallet, slots, roulette, plinko, blackjack, social.
- `frontend/src/pages/casino-miniapp-page/ui/CasinoMiniAppPage.tsx` и CSS — стоит разделить на shell/state hooks/panels.
- importers в `tools/importers/*/app/main.go` — можно выделять parser/service/writer слои по мере изменений.

Безопасная следующая декомпозиция:

1. `store_roulette.go` — roulette state/bets/history/settlement.
2. `store_player.go` — ensurePlayer, balance/wallet state, profile.
3. `casinoApi` endpoint injection split на feature API slices.
4. Casino mini app shell: вынести tab state, profile modal state и game panels.

## Горящие риски

- Реальные секреты уже должны считаться скомпрометированными до внешней ротации.
- Prod compose теперь правильно fail-fast, но VPS env и DB grants должны быть подготовлены руками.
- Локальная демка требует фактического browser smoke, потому что build/test не подтверждают UX и route refresh.
- `go test ./...` всё ещё видит Go-пакеты внутри `node_modules` зон; это devex-долг, но сейчас не блокирует тесты.
- `cmd/*` stubs и legacy compose остаются transitional шумом; не использовать их как MVP entrypoints.

## MVP критерии готовности

Минимальный внутренний показ:

- `make core-up && make migrate-runtime`;
- `make casino-up`;
- `cd frontend && npm.cmd run dev`;
- `/api/front?limit=1` непустой;
- login/register/me работают;
- `/casino`: balance + одно действие + history;
- `/admin`: виден только admin user;
- mobile 375px/640px без скрытой навигации.

Публичный pre-MVP:

- все секреты ротированы;
- HTTPS/HSTS включены;
- DB роли non-superuser;
- migrations up-only применены;
- CI зелёный;
- smoke по основным маршрутам зафиксирован.
