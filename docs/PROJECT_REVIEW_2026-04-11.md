# Project Review — mudro
**Дата:** 2026-04-11
**Ветка:** `agent/20260405-project-review-pass` (есть незакоммиченные правки в казино/рулетке)
**Last main commit:** `4eb3aa1 docs: add casino roadmap 2026` (2026-04-05)

---

## 1. Актуальные ветки

Всё живёт вокруг даты **2026-04-05**, после этого фиксированных коммитов в `origin/main` нет.

| Ветка | Назначение | Статус |
|---|---|---|
| `main` | текущий релиз v0.1.0-mvp | stable, HEAD 4eb3aa1 |
| `agent/20260405-project-review-pass` | **текущая рабочая**, fairness + prod hardening | впереди main, есть `modified:` в рулетке |
| `agent/20260404-casino-stack` | казино + admin + openrouter | уже влит в main через `claude/brave-aryabhata` |
| `agent/20260403-pre-mudro-mvp` | MVP hardening, Blackjack, dark UI | влит |
| `agent/20260401-logger` | Vercel deploy fixes | отдельный трек, не в main |

**Мусор:** ~15 веток `feat/random-circle*`, `feature/random-circles-*`, `claude/*`, `codex/*`, `tmp-tree-test-*`, `probe-tree-sha-*` — это экспериментальные/worktree-ветки, которые пора почистить. Репо в `.git/index` локально **битый** (`bad signature 0x00000000`) — нужно `git reset` или заново клонировать.

**Немедленно:** закоммитить или откатить 3 грязных файла на `agent/20260405-project-review-pass`:
- `frontend/src/features/casino/ui/RouletteWheelSVG.tsx`
- `services/casino/internal/casino/roulette.go`
- `services/casino/internal/casino/store.go`

---

## 2. Структура репо

Go 1.24, React 19, Vite 7, pgx/v5, Redis, Kafka (redpanda), MinIO.

```
services/        # активные раннеры
  feed-api/      # главный HTTP, ~90 роутов /api/*, mux из server.go
  casino/        # отдельный сервис + своя БД, 11 миграций, store.go 2149 строк
  agent/         # фоновая обработка
  bot/           # telegram, openrouter
  auth-api/      # additive, в процессе миграции
  bff-web/       # additive
  api-gateway/   # additive, проксирует в feed/bff/auth/orchestration
  orchestration-api/
  movie-catalog/

internal/        # shared-пакеты: auth, posts, chat, media, ratelimit, config, bot, ...
tools/           # CLI: importers/{vk,tg,...}, backfill, maintenance, opus-gateway
cmd/             # legacy-раннеры, ещё используются (backfill, memento, mudro.exe)
migrations/      # 27 up-миграций main-контура + movie_catalog/
services/casino/migrations/  # 10 up-миграций казино
ops/compose/     # canonical docker-compose: core + services + legacy
docs/            # много ADR/runbook/иногда протухшие доки
```

**Проблема:** параллельно живут `cmd/*` (legacy) и `services/*/cmd` (новый формат). Миграция заявлена как «в процессе» ещё с марта. Пора закрывать — либо удалить `cmd/*`, либо перенести те части, что ещё нужны, в `tools/`.

---

## 3. Backend

### feed-api
`services/feed-api/app/run.go` → `wiring.NewHandler` → `feed.Server.Router()`.

`feed/server.go` регистрирует **~90 роутов через `net/http.ServeMux`** вручную, одной кучей. Все казино-эндпоинты (`/api/casino/**`) висят на `feed-api` как **прокси** к отдельному casino-сервису (`casino_proxy.go`, ~15 ручек). Это антипаттерн — fanout через monolith, но при этом ещё и api-gateway, который формально должен это делать.

**Плюсы:** готовый graceful shutdown, опциональный Redis rate limiter (fixed window), CORS, JWT, базовая ratelimit из `internal/ratelimit`.

**Минусы / долги:**
- `server.go` — 175 строк декларации роутов, пора переходить на что-то с method dispatch (хотя бы Go 1.22+ `mux.HandleFunc("GET /...")`, он уже доступен).
- Двойной стек auth: `internal/auth` + отдельный `services/auth-api`. Реально работает первый, второй — скелетный additive.
- API-gateway существует но не в продакшн-пайплайне — трафик идёт напрямую в feed-api.
- `casino_proxy.go` делает ~30 методов вручную, нет generic-прокси.
- Три миграции в 009-009c и 011-011c — признак того, что дата-модель переделывалась на ходу под недостроенный социальный слой.

### casino
Отдельный deploy, своя БД, свой DSN. `store.go` **2149 строк** — это God-object. Внутри: rtp, round lifecycle, bonus, blackjack, plinko, roulette, wallet sync, audit. Пора бить на пакеты.

Плюсы: fairness module (`fairness.go`, `019_unified_fairness.sql`), wallet-sync через `CASINO_MAIN_DSN`, idempotency-таблица, ledger-entries (двойная запись). По duck-test — денежный контур сделан аккуратно.

Минусы:
- `store.go` 2k строк и `handlers.go` 842 строки — монолит.
- В `agent/20260405-*` незакоммичено — значит правки в проде fairness/рулетки не задокументированы.

### Остальные сервисы
`auth-api`, `bff-web`, `api-gateway`, `orchestration-api`, `movie-catalog` — у всех есть `cmd/main.go`, `internal/*/handler.go`, тесты. По сути это тонкие каркасы, которые в прод пока не включены: в `ops/compose/docker-compose.services.yml` они поднимаются, но в `docker-compose.prod.yml` из них присутствует только `casino-api`. **То есть микросервисная раскладка — чертёж, а не прод.**

---

## 4. Frontend

FSD: `app/ entities/ features/ pages/ shared/ widgets/`. **93 файла, ~11k строк**, 6 тестов. Redux Toolkit + RTK Query с единым `mudroApi` (`shared/api/mudroApi.ts`) и `injectEndpoints` по доменам:

```
entities/post/model/postsApi.ts
entities/session/api/authApi.ts
entities/chat/api/chatApi.ts
features/casino/api/casinoApi.ts
features/admin/api/adminApi.ts
features/orchestration/api/orchestrationApi.ts
```

Корректно реализован `baseQueryWithReauth` (401 → `/auth/refresh` → retry), JWT в заголовке из `session.token`. `env.apiBaseUrl` из `VITE_API_BASE_URL`, фоллбек `/api`. `RootState` типизирован нормально, `AuthInitializer` дергает `useMeQuery` при наличии токена.

**Страницы:** feed, casino (отдельно miniapp), chat, movies + movie-catalog, admin, orchestration, auth (login/register/profile). **TODO/FIXME в коде — 0.**

**Замечания:**
- React 19, Vite 7, Tailwind 4 — всё свежее. Отдельно `@tailwindcss/vite` 4.2 — ок.
- `lucide-react@0.577` и `framer-motion@12` — свежие, риски breaking minor.
- Тестов — **6 файлов** на 93 исходника. Критические потоки (казино, feed, auth) не покрыты unit'ами.
- Нет `msw` / storybook — визуальная регрессия не ловится.

---

## 5. Тесты и CI

### Покрытие
| Зона | Test files | Оценка |
|---|---|---|
| `internal/` (Go shared) | 14 (auth, bot, catalog, chat, config, events, media, ratelimit, reporter) | **средне** — покрыто главное |
| `services/` (новые) | 12 (auth-api, bff-web, casino, feed-api, movie-catalog, orchestration, api-gateway) | **тонко** — большинство это 1 handler_test.go на сервис |
| `tools/importers` + maintenance | 9 | нормально |
| `frontend/src` | 6 | **мало** |

Всего Go `*_test.go`: **50** (в рабочем дереве, без `.claude/worktrees`).

### selftest
```make
selftest:
	$(GO) test ./tools/importers/vkimport/app ./tools/importers/tgimport/app
```
То есть `make selftest` проверяет **только 2 импортера**. Это совсем не «быстрый прогон проекта». Для реального selftest нужно гонять хотя бы `./services/... ./internal/...` (и выкинуть интеграционные за тег).

### CI
`.github/workflows/ci.yml` — **3 job'а**:
1. `security` — gitleaks.
2. `backend` — `golangci-lint` + `go test -v ./...` (триггерит ВСЕ тесты, включая integration — это дорого и flaky).
3. `frontend` — `npm install` (не `ci`) + lint + build.

**Чего нет:**
- `test-backend` не разделён на unit и integration.
- Frontend jobs не гоняет `npm test` / vitest.
- Нет job'а на `go build ./services/...` отдельно от тестов — если тест-пакет сломан, падает билд всех.
- Нет smoke-e2e, нет `validate-microservices` (которые были в memory — с тех пор CI ужали).
- `npm install` вместо `npm ci` — недетерминированный фронт-build.
- `actions/setup-node@v4`, Node 20 — ок.

---

## 6. DevOps и деплой

### Раскладка
- **Prod-таргет:** VPS RU17840 (91.218.113.247), копия в `/srv/mudro`.
- **Dockerfile** — один multi-stage с `ARG SERVICE=services/feed-api/cmd`, `distroless/static:nonroot`. Чисто и правильно.
- **docker-compose.yml** (в корне) — **конфликтует** с `docker-compose.prod.yml`:
  - корневой: `postgres_main` (5432) + `postgres_casino` (5433), credentials `mudro/mudropass`, БД `mudro_main`/`mudro_casino`.
  - prod: `db` (5433) + `casino-db` (5434), БД `gallery`/`mudro_casino`, user `postgres`.
  - Это **два разных схемных мира**. Корневой compose вообще никто не использует — удалить или пометить deprecated.
- **ops/compose/docker-compose.core.yml** — канонический, запускается через `make core-up`. Поднимает db/redis/kafka(redpanda).
- `ops/compose/docker-compose.services.yml` — монтирует исходники (`..:/app`) и гоняет `go run` — это **dev-режим**, не прод. В prod должен собираться Dockerfile.

### Makefile
10+ kB, куча целей: `core-up`, `micro-up`, `migrate-runtime`, `test-active`, `health`, `health-runtime`, `count-posts-core` и т.д. В целом работает, но:
- Внутри дублирование списков миграций.
- `RUNTIME_MIGRATIONS` захардкожен и не включает 010, 011*, 012*, 013-019 — т.е. `make migrate-runtime` **не прокатит всю схему**. Это скрытый баг.
- `selftest` слабый (см. выше).

### Systemd / nginx
`ops/systemd/`, `ops/nginx/` — есть, конфиги заготовлены (nginx с HTTPS-редиректом и SSL placeholders по CHANGELOG). В ревью содержимого не лез.

### Секреты
`.env.example` покрыт, `.env` игнорится, правило «секреты вне repo, 600» соблюдается. gitleaks в CI есть. В репо валяются `mudro.exe` (18MB) и `casino.test.exe` (8.3MB) — **в .gitignore они есть**, но физически на диске остаются и могут случайно уйти в PR если кто-то снимет маску. Удалить.

### Оценка готовности к DevOps
**6/10.** Базис есть (Dockerfile, compose, migrations-скрипты, systemd, CI, gitleaks), но:
- микросервисный compose это dev-режим (`go run`), нет production-картинок в registry;
- два конфликтующих compose в корне;
- CI гоняет всё подряд, нет разделения unit/integration;
- `make migrate-runtime` не применяет половину миграций — риск для любого нового окружения;
- нет явного healthcheck-пайплайна "после деплоя";
- нет container registry/image-tag стратегии;
- нет observability (metrics/traces) в коде — только логи.

---

## 7. База данных

### Главный контур (`migrations/`)
27 up-миграций, bigserial PK, 30+ таблиц. Домены:
- **posts/reactions/comments/media** — ядро ленты, с историей переделок (008, 010, 011b, 013).
- **users/auth** — 009 → 011 (simplify) → 011b (relationships) → 012b (finalize). Дублирующаяся эволюция схемы.
- **follows/conversations/messages/notifications** — социальный слой (`010c_social_messenger_agents.sql`, 7.4 kB).
- **agents** + `agent_queue`, `agent_task_events`, `agent_review_gate` — внутренняя оркестрация.
- **chat** (016), stickers, chat_messages.
- **casino** (012, 017, 018, 019 — unified_fairness).

Экстеншены: **только `pgcrypto`** (в 012). Триггеры: `014_likes_count_trigger`.

### Казино контур (`services/casino/migrations/`)
10 миграций: init → live roulette → bonus → live social → activity indexes → projection sync → blackjack → constraints → bet constraints → performance indexes → unified fairness. **Это параллельная БД** с кроссконтурной синхронизацией кошелька через `CASINO_MAIN_DSN`.

### Риски данных
- Нет down-миграций для `019_unified_fairness.sql` (есть только .sql, нет .down.sql).
- `011b_fix_relationships.sql` фактически переделывает foreign-key граф — опасно при откате.
- Нет init-sql для `movie_catalog/` в `RUNTIME_MIGRATIONS` (есть, но только один файл).

---

## 8. Готовность к Supabase

Реальный след Supabase в репо минимален:
- `.env.example`: `# For Supabase-backed casino, replace CASINO_DSN with ... sslmode=require`.
- `README.md`: абзац про Supabase-backed casino.
- `docs/casino-db-stabilization-checklist.md`: галочка про `sslmode=require`.

То есть заготовлено только для казино-контура, и только как "переключить DSN". Для полного переезда на Supabase есть **жёсткие блокеры**:

### Блокеры

1. **RLS отсутствует полностью.** В миграциях 0 матчей `ROW LEVEL SECURITY` / `auth.uid()` / `alter table ... enable row level security`. Supabase без RLS — это дыра, потому что anon-key в браузере без политик открывает всю БД. Либо ты пускаешь клиентов только через свой Go-бэкенд (и тогда Supabase используется просто как managed Postgres — но тогда не нужен Supabase), либо тебе придётся написать RLS-политики на весь граф (users/posts/comments/messages/casino_*) — это неделя-две работы.

2. **Свой JWT + свой `users`.** Таблица `users` с `bigserial PK` + `auth_tokens` + `user_sessions` + свой `JWTSecret`. Supabase Auth это `auth.users (uuid PK)`, JWK-подписи, триггеры. Для переезда надо:
   - либо оставить свою auth и игнорить Supabase Auth (тогда никакого клиентского SDK),
   - либо переложить `users` на `auth.users`, поменять PK на `uuid`, переписать FK по всей схеме, мигрировать данные и переписать `internal/auth`.

3. **`bigserial PK` vs `uuid`.** Supabase-stack рассчитан на uuid (realtime, RLS-policies по `auth.uid()`). Переход — ломающая миграция всех FK.

4. **Realtime / WebSocket.** Чат и рулетка сейчас на самописных WS-хабах (`internal/chat/hub.go`, `services/casino/...stream`). Supabase Realtime на pg-replication-logical — другой контракт. Либо оставляем свои хабы (и тогда Supabase не нужен для этой фичи), либо переписываем на Supabase channels.

5. **Кроссконтурная транзакция wallet-sync.** `CASINO_MAIN_DSN` — казино коннектится **в чужую БД** для списания баланса кошелька в main. В Supabase это две разные проект-базы без общей транзакции. Надо переделывать на saga/outbox, иначе денежный инвариант ломается.

6. **Два compose/две БД.** У тебя main и casino разделены физически. Supabase-план подразумевает **один проект = одна БД**. Либо объединять, либо заводить два Supabase-проекта (и всё равно переделывать wallet-sync).

7. **Migration tooling.** Сейчас миграции катятся через `psql` из Makefile. Для Supabase CLI нужен другой формат (`supabase/migrations/*.sql` с таймстемпами). Это не блокер, но переупаковка.

### Что реально готово для Supabase
- `pgcrypto` уже подключен.
- `pgx/v5` спокойно ходит на `sslmode=require` — серверная сторона зайдёт без правок.
- Индексы расставлены, таблицы нормализованы, `created_at/updated_at` есть.

### Моя рекомендация
Если цель — **managed Postgres** (не Supabase Auth / Realtime / SDK), то Supabase не нужен вообще. Возьми Neon/Timescale/любой managed Postgres и просто поменяй `DATABASE_URL`. Это одна ночь работы.

Если цель — **уйти от VPS и отдавать фронт напрямую в Supabase** (anon-key, клиентский SDK, RLS), то это **пере-архитектура auth и моделей пользователя на 2-4 недели**, с миграцией PK на uuid и написанием политик. Сейчас не готово.

---

## 9. Что делать в ближайшие циклы

Короткий список, приоритет сверху вниз:

1. **Почистить ветки и index.** Локальный `.git` битый. Закоммитить/откатить незакоммиченные изменения в казино. Удалить 15+ experimental-веток. Разово привести `main`/`agent/*` в порядок.
2. **Починить `RUNTIME_MIGRATIONS`** в Makefile — включить все миграции 010-019. Сейчас `make migrate-runtime` оставляет схему незаконченной.
3. **Разнести CI** backend на unit+integration (build-теги или `-short`), добавить `npm test` + `npm ci`.
4. **Усилить `selftest`** — прогон `go test -short ./services/... ./internal/...` за <2 минуты. Текущая версия проверяет 2 импортера.
5. **Разбить casino/store.go** (2149 строк) на `store/wallet.go`, `store/rounds.go`, `store/rtp.go`, `store/bonus.go`, `store/games_*.go`. Без изменения контрактов.
6. **Удалить** корневой `docker-compose.yml` или пометить deprecated; оставить единственный канонический в `ops/compose/`.
7. **Удалить** `cmd/*` legacy или явно пометить что это tools-only и перенести в `tools/`.
8. **Покрытие фронта** — хотя бы по 1 тесту на каждую страницу + RTK Query endpoints (cassino, auth, posts).
9. **ADR на Supabase** — зафиксировать решение "managed Postgres vs full Supabase stack" в `docs/adr/`, иначе реально потратите месяц на неправильный план.
10. **Sanity-check `mudro.exe` и `casino.test.exe`** — удалить с диска, не только из git-индекса.

---

## 10. TL;DR

- **Main живой**, последний коммит 2026-04-05, текущий рабочий фронт — `agent/20260405-project-review-pass` с грязным деревом.
- **Монорепо в процессе миграции** cmd→services уже месяц — надо закрывать.
- **Backend — 60+ ручек в одном feed-api mux**, казино отдельный сервис с God-object-ом в 2k строк. Работает, но чинить надо.
- **Frontend — ОК**, FSD+RTK Query правильно разложены, только тестов мало.
- **CI жиденький**, `selftest` — фасад на 2 импортера.
- **DevOps — база есть**, но два конфликтующих compose, `make migrate-runtime` ломаный, observability нет.
- **БД — bigserial, без RLS**, два физических контура, wallet-sync через кроссконтурное подключение.
- **Supabase готовность — только для казино-DSN**. Полный переезд требует переделки auth, PK-типов и wallet-sync. Если нужен просто managed Postgres — Supabase не нужен.
