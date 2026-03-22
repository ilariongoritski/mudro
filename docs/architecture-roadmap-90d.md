# MUDRO: 90-дневный план архитектуры (2026-03-22 → 2026-06-20)

**Контекст**: Go-проект социальной ленты (VK+TG), 36 таблиц БД, 2687 постов в проде, React frontend на Vercel, API на VPS Ubuntu 24.04. Референсы: usememos/memos, bluesky-social/indigo, tinode/chat, GoToSocial.

---

## 1. Текущее состояние (baseline 2026-03-22)

### Структура репозитория
```
mudro11-main/
├── cmd/
│   ├── api/          # HTTP API (pgxpool, auth, posts service)
│   ├── bot/          # Telegram bot (команды /health, /feed5, /mudro)
│   ├── agent/        # Автономный worker (planner + executor)
│   ├── reporter/     # Периодические отчёты в TG
│   └── *import/      # 9 импортёров (vk, tg, comments, media)
├── internal/
│   ├── api/          # HTTP handlers (auth, admin, feed)
│   ├── auth/         # JWT service
│   ├── posts/        # Posts service (LoadPosts, reactions, comments)
│   ├── agent/        # Planner/Worker для задач
│   ├── bot/          # TG bot logic
│   ├── media/        # Media normalization
│   ├── ratelimit/    # Token bucket + Redis limiter
│   └── events/       # Kafka publisher (опционально)
├── migrations/       # 11 SQL миграций (001-011)
├── frontend/         # React+TS+RTK+Vite (FSD layout)
├── docker-compose.prod.yml  # db, redis, kafka, api, agent, reporter, minio
└── scripts/mcp/      # MCP-серверы (filesystem, git, postgres, github)
```

### Стек
- **Backend**: Go 1.24, pgx/v5, JWT, HTTP stdlib
- **DB**: Postgres 16, 36 таблиц, 58 FK
- **Frontend**: React 18, TypeScript, Redux Toolkit, RTK Query, Vite
- **Инфра**: Docker Compose, Redis, Kafka (Redpanda), MinIO, Vercel (frontend), VPS (API)
- **Боты**: Telegram bot-api/v5
- **MCP**: filesystem, git, postgres (read-only), github (опционально)

### Ключевые таблицы
- `posts` (2687 строк: vk=1088, tg=1599)
- `post_comments` (1806 комментариев)
- `post_reactions`, `comment_reactions`
- `media_assets`, `post_media_links`, `comment_media_links`
- `users`, `user_roles`, `user_subscriptions`
- `agent_queue`, `agent_task_events`, `agent_review_gate`
- `account_post_likes`

### Что работает
✅ API `/api/posts`, `/api/front`, `/feed` (HTML+JSON)
✅ Auth `/api/auth/register`, `/api/auth/login`, `/api/auth/me`
✅ Admin `/api/admin/users`, `/api/admin/stats`
✅ TG bot команды: `/health`, `/feed5`, `/mudro`, `/time`, `/agent24`
✅ Agent worker (planner + executor, 15s цикл)
✅ Reporter (30min отчёты в TG)
✅ Frontend на Vercel (feed page, detail drawer, filters)
✅ Media backfill (1591/1778 HTTP URLs)
✅ Rate limiting (token bucket + Redis)
✅ Kafka events (опционально)

### Что НЕ работает / не готово
❌ HTTPS на VPS (только HTTP :80)
❌ MinIO bucket не настроен (сервис есть, backfill не выполнен)
❌ Sticker packs + custom emoji (таблицы есть, UI нет)
❌ Casino user_id миграция (BIGINT FK готов, но не протестирован)
❌ Legacy jsonb колонки (posts.media, comments.media) не мигрированы на *_media_links
❌ GitHub PAT не заполнен (MCP падает)
❌ Admin panel UI (нет экрана)
❌ Mobile polish (desktop OK, mobile не проверен)
❌ Login/Register UI (светлые стили, не dark glassmorphism)

---

## 2. Референсные проекты: что копировать, что НЕ копировать

### 2.1 usememos/memos (Go + React, личные заметки)
**Копировать концептуально**:
- Минималистичный HTTP API (stdlib, без фреймворков)
- Встроенный SQLite для dev, Postgres для prod
- Единый бинарник с embedded frontend (Go 1.16+ embed)
- Markdown-first контент (у нас — text + media)
- Простая auth (JWT, без OAuth сложности)
- Webhook-интеграции (у нас — Telegram bot)

**НЕ копировать**:
- SQLite в проде (у нас уже Postgres)
- Отсутствие микросервисов (у нас agent/reporter уже отдельно)
- Отсутствие реакций/комментариев (у нас есть)

**Применить к MUDRO**:
- Упростить API: убрать лишние middleware, оставить stdlib + pgx
- Рассмотреть embed frontend в Go бинарник (альтернатива Vercel)
- Добавить webhook для внешних интеграций (Discord, Slack)

### 2.2 bluesky-social/indigo (Go, AT Protocol, федеративная соцсеть)
**Копировать концептуально**:
- Federated identity (DID, у нас — локальные users + JWT)
- Event streams (firehose, у нас — Kafka events)
- Lexicon schema (типизированные события, у нас — agent tasks)
- Moderation API (у нас — admin panel)
- Blob storage (у нас — MinIO)

**НЕ копировать**:
- AT Protocol сложность (DID, PDS, BGS, AppView)
- Федерация (у нас централизованный backend)
- IPLD/CAR форматы (у нас Postgres + JSON)

**Применить к MUDRO**:
- Kafka event schema: `posts.created`, `comments.created`, `reactions.added`
- Blob storage: MinIO для media (вместо inline URLs)
- Moderation queue: `agent_queue` для review-задач
- Firehose endpoint: `/api/firehose` (SSE stream новых постов)

### 2.3 tinode/chat (Go, real-time чат)
**Копировать концептуально**:
- WebSocket для real-time (у нас — SSE или WS для live feed)
- Presence (online/offline, у нас — user activity)
- Push notifications (у нас — Telegram bot)
- Message threading (у нас — comment threads)
- Read receipts (у нас — views_count)

**НЕ копировать**:
- Собственный протокол (у нас REST + JSON)
- gRPC (у нас HTTP)
- Отсутствие SQL (у нас Postgres)

**Применить к MUDRO**:
- WebSocket endpoint `/api/ws` для live feed updates
- Presence API: `GET /api/users/:id/presence`
- Push notifications через Telegram bot (уже есть)
- Comment threading UI (уже есть `parent_comment_id`)

### 2.4 GoToSocial (Go, ActivityPub, Mastodon-совместимый)
**Копировать концептуально**:
- ActivityPub federation (у нас — экспорт в RSS/Atom)
- Media processing (resize, thumbnails, у нас — media_assets)
- Content warnings (у нас — moderation flags)
- Custom emoji (у нас — sticker packs)
- Account migration (у нас — export/import)

**НЕ копировать**:
- ActivityPub сложность (у нас централизованный backend)
- Федерация (у нас один инстанс)
- Mastodon API совместимость (у нас свой API)

**Применить к MUDRO**:
- RSS/Atom feed: `/api/feed.rss`, `/api/feed.atom`
- Media processing: thumbnails для preview_url
- Custom emoji: UI для sticker packs
- Export API: `GET /api/users/:id/export` (JSON archive)

---

## 3. 90-дневный план (3 спринта по 30 дней)

### Спринт 1 (2026-03-22 → 2026-04-20): Стабилизация + Real-time

**Цель**: Закрыть P0/P1 долги, добавить WebSocket live feed, HTTPS на VPS.

**Workstreams**:

#### WS1.1: Security & Ops (P0)
- [ ] Ротация Telegram/OpenAI токенов (засвечены в git)
- [ ] HTTPS на VPS (Let's Encrypt + Nginx)
- [ ] MinIO bucket setup + backfill URLs в `media_assets`
- [ ] GitHub PAT для MCP (fine-grained, read-only)
- [ ] Firewall audit (UFW rules, fail2ban)

**Референс**: GoToSocial (media storage), memos (простой deploy)

#### WS1.2: Real-time Feed (WebSocket)
- [ ] WebSocket endpoint `/api/ws` (stdlib `net/http`)
- [ ] SSE fallback `/api/events` (для старых браузеров)
- [ ] Kafka consumer → WS broadcast (новые посты/комментарии)
- [ ] Frontend: RTK Query + WS subscription
- [ ] Presence API: `GET /api/users/:id/presence`

**Референс**: tinode (WebSocket), bluesky (firehose)

#### WS1.3: Mobile Polish + UI
- [ ] Mobile responsive (feed page, drawer, toolbar)
- [ ] Login/Register dark glassmorphism
- [ ] Admin panel UI (users list, stats, agent queue)
- [ ] Sticker packs UI (emoji picker)

**Референс**: memos (минималистичный UI)

#### WS1.4: Testing & CI
- [ ] Unit tests для `internal/api`, `internal/posts`
- [ ] Integration tests для WebSocket
- [ ] GitHub Actions: `go test`, `npm run build`, `docker build`
- [ ] E2E smoke test (Playwright)

**Референс**: все проекты (CI/CD best practices)

**Deliverables**:
- ✅ HTTPS на VPS
- ✅ WebSocket live feed
- ✅ Mobile UI polish
- ✅ CI pipeline

---

### Спринт 2 (2026-04-21 → 2026-05-20): Federation + Export

**Цель**: RSS/Atom feed, export API, media processing, custom emoji.

**Workstreams**:

#### WS2.1: Federation (RSS/Atom)
- [ ] RSS feed: `/api/feed.rss` (XML, последние 50 постов)
- [ ] Atom feed: `/api/feed.atom` (XML, последние 50 постов)
- [ ] Per-user feed: `/api/users/:id/feed.rss`
- [ ] Frontend: RSS icon + subscribe link

**Референс**: GoToSocial (ActivityPub), memos (webhook)

#### WS2.2: Export API
- [ ] Export endpoint: `GET /api/users/:id/export` (JSON archive)
- [ ] Export format: posts + comments + media + reactions
- [ ] Import endpoint: `POST /api/users/:id/import` (JSON archive)
- [ ] Frontend: export button в profile

**Референс**: GoToSocial (account migration), memos (backup)

#### WS2.3: Media Processing
- [ ] Thumbnail generation (resize to 300x300, 600x600)
- [ ] Preview URL backfill в `media_assets`
- [ ] Image optimization (WebP, AVIF)
- [ ] Video thumbnails (ffmpeg)

**Референс**: GoToSocial (media processing), bluesky (blob storage)

#### WS2.4: Custom Emoji
- [ ] Sticker packs UI (upload, manage)
- [ ] Emoji picker в comment form
- [ ] Emoji reactions (вместо text reactions)
- [ ] Emoji search (по keyword)

**Референс**: GoToSocial (custom emoji)

**Deliverables**:
- ✅ RSS/Atom feed
- ✅ Export/Import API
- ✅ Media thumbnails
- ✅ Custom emoji UI

---

### Спринт 3 (2026-05-21 → 2026-06-20): Microservices + Scale

**Цель**: Декомпозиция на микросервисы, Kafka event backbone, horizontal scaling.

**Workstreams**:

#### WS3.1: Microservices Split
- [ ] Выделить `services/feed-api` (HTTP API для постов)
- [ ] Выделить `services/agent-planner` (планировщик задач)
- [ ] Выделить `services/agent-worker` (исполнитель задач)
- [ ] Выделить `services/telegram-bot` (TG bot)
- [ ] Выделить `services/reporter` (отчёты)

**Референс**: bluesky (AppView, PDS, BGS), tinode (микросервисы)

#### WS3.2: Kafka Event Backbone
- [ ] Kafka topics: `posts.created`, `comments.created`, `reactions.added`, `tasks.created`, `tasks.done`
- [ ] Producer в `feed-api` (новые посты → Kafka)
- [ ] Consumer в `agent-planner` (задачи → Kafka)
- [ ] Consumer в `telegram-bot` (уведомления → Kafka)
- [ ] Consumer в `reporter` (метрики → Kafka)

**Референс**: bluesky (firehose), tinode (event streams)

#### WS3.3: Horizontal Scaling
- [ ] Redis для distributed rate limiting (уже есть)
- [ ] Redis для session storage (вместо JWT в cookie)
- [ ] Postgres read replicas (для `/api/posts`)
- [ ] Load balancer (Nginx upstream)
- [ ] Health checks для всех сервисов

**Референс**: все проекты (production best practices)

#### WS3.4: Monitoring & Observability
- [ ] Prometheus metrics (`/metrics` endpoint)
- [ ] Grafana dashboards (API latency, DB queries, Kafka lag)
- [ ] Structured logging (JSON logs)
- [ ] Distributed tracing (OpenTelemetry)

**Референс**: все проекты (observability)

**Deliverables**:
- ✅ Микросервисы (5 сервисов)
- ✅ Kafka event backbone
- ✅ Horizontal scaling
- ✅ Monitoring

---

## 4. Что НЕ копировать (антипаттерны)

### Из usememos/memos
❌ SQLite в проде (у нас Postgres)
❌ Отсутствие микросервисов (у нас уже agent/reporter)
❌ Отсутствие реакций/комментариев (у нас есть)

### Из bluesky-social/indigo
❌ AT Protocol сложность (DID, PDS, BGS, AppView)
❌ Федерация (у нас централизованный backend)
❌ IPLD/CAR форматы (у нас Postgres + JSON)

### Из tinode/chat
❌ Собственный протокол (у нас REST + JSON)
❌ gRPC (у нас HTTP)
❌ Отсутствие SQL (у нас Postgres)

### Из GoToSocial
❌ ActivityPub сложность (у нас централизованный backend)
❌ Федерация (у нас один инстанс)
❌ Mastodon API совместимость (у нас свой API)

---

## 5. Highest-leverage Go tooling & integrations

### 5.1 Добавить СЕЙЧАС (Спринт 1)

#### golangci-lint (линтер)
```bash
# .golangci.yml
linters:
  enable:
    - gofmt
    - govet
    - staticcheck
    - errcheck
    - gosimple
    - ineffassign
    - unused
```

**Зачем**: Автоматический code review, ловит 80% багов до runtime.

#### air (hot reload)
```bash
# .air.toml
[build]
  cmd = "go build -o ./tmp/main ./cmd/api"
  bin = "./tmp/main"
  include_ext = ["go", "tpl", "tmpl", "html"]
```

**Зачем**: Перезапуск API при изменении кода (dev experience).

#### sqlc (type-safe SQL)
```bash
# sqlc.yaml
version: "2"
sql:
  - schema: "migrations/*.sql"
    queries: "internal/db/queries.sql"
    engine: "postgresql"
    gen:
      go:
        package: "db"
        out: "internal/db"
```

**Зачем**: Генерация Go кода из SQL (вместо ручного `Scan`).

**Референс**: все проекты используют sqlc или аналоги.

#### testcontainers-go (integration tests)
```go
postgres, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
    ContainerRequest: testcontainers.ContainerRequest{
        Image: "postgres:16",
        Env:   map[string]string{"POSTGRES_PASSWORD": "test"},
    },
})
```

**Зачем**: Реальный Postgres в тестах (вместо моков).

**Референс**: bluesky, GoToSocial используют testcontainers.

### 5.2 Добавить в Спринте 2

#### oapi-codegen (OpenAPI → Go)
```bash
oapi-codegen -package api -generate types,server openapi.yaml > internal/api/generated.go
```

**Зачем**: Генерация API handlers из OpenAPI spec.

**Референс**: memos использует OpenAPI для документации.

#### go-swagger (Swagger UI)
```bash
swagger serve ./docs/swagger.yaml
```

**Зачем**: Интерактивная документация API.

#### migrate (миграции)
```bash
migrate -path migrations -database "postgres://..." up
```

**Зачем**: Управление миграциями (уже используем, но можно улучшить).

### 5.3 Добавить в Спринте 3

#### Prometheus client (метрики)
```go
import "github.com/prometheus/client_golang/prometheus"

httpRequestsTotal := prometheus.NewCounterVec(
    prometheus.CounterOpts{Name: "http_requests_total"},
    []string{"method", "endpoint", "status"},
)
```

**Зачем**: Метрики для Grafana.

**Референс**: все проекты используют Prometheus.

#### OpenTelemetry (tracing)
```go
import "go.opentelemetry.io/otel"

tracer := otel.Tracer("mudro")
ctx, span := tracer.Start(ctx, "LoadPosts")
defer span.End()
```

**Зачем**: Distributed tracing для микросервисов.

**Референс**: bluesky, tinode используют tracing.

#### Temporal (workflow engine)
```go
import "go.temporal.io/sdk/workflow"

func ImportWorkflow(ctx workflow.Context, source string) error {
    // multi-step import: download → parse → validate → insert
}
```

**Зачем**: Долгие задачи (import, export, media processing) как workflows.

**Референс**: альтернатива agent_queue для сложных задач.

---

## 6. Repo integrations (GitHub)

### 6.1 GitHub Actions (CI/CD)
```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      - run: go test ./...
      - run: golangci-lint run
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: docker build -t mudro-api .
```

**Зачем**: Автоматические тесты + build на каждый push.

### 6.2 Dependabot (обновление зависимостей)
```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
  - package-ecosystem: "npm"
    directory: "/frontend"
    schedule:
      interval: "weekly"
```

**Зачем**: Автоматические PR для обновления Go modules + npm packages.

### 6.3 CodeQL (security scanning)
```yaml
# .github/workflows/codeql.yml
name: CodeQL
on: [push, pull_request]
jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: github/codeql-action/init@v2
        with:
          languages: go, javascript
      - uses: github/codeql-action/analyze@v2
```

**Зачем**: Автоматический поиск уязвимостей.

### 6.4 GitHub Projects (roadmap)
- Создать Project board для 90-дневного плана
- Колонки: Backlog, Sprint 1, Sprint 2, Sprint 3, Done
- Issues для каждого workstream

**Зачем**: Прозрачный roadmap для команды.

---

## 7. Критические решения (trade-offs)

### 7.1 Monolith vs Microservices
**Решение**: Начать с monolith (Спринт 1-2), split в Спринте 3.

**Обоснование**:
- Референсы (memos, GoToSocial) начинали с monolith
- Микросервисы добавляют сложность (network, deployment)
- У нас уже есть agent/reporter (частичный split)

**План**:
- Спринт 1-2: monolith API + agent + reporter
- Спринт 3: split на feed-api, agent-planner, agent-worker, bot, reporter

### 7.2 REST vs GraphQL
**Решение**: REST (текущий подход).

**Обоснование**:
- Все референсы используют REST (кроме bluesky AT Protocol)
- GraphQL добавляет сложность (schema, resolvers)
- REST проще для mobile clients

**Альтернатива**: Добавить GraphQL endpoint в Спринте 3 (опционально).

### 7.3 JWT vs Session
**Решение**: JWT (текущий подход) → Session в Спринте 3.

**Обоснование**:
- JWT проще для stateless API
- Session лучше для security (revocation, rotation)
- Референсы (tinode, GoToSocial) используют session

**План**:
- Спринт 1-2: JWT
- Спринт 3: Redis session storage

### 7.4 Postgres vs NoSQL
**Решение**: Postgres (текущий подход).

**Обоснование**:
- Все референсы используют SQL (Postgres, SQLite)
- Relational data (posts, comments, reactions)
- ACID гарантии

**Альтернатива**: Redis для cache, но не для primary storage.

### 7.5 Vercel vs Self-hosted Frontend
**Решение**: Vercel (текущий подход) → Self-hosted в Спринте 2.

**Обоснование**:
- Vercel проще для MVP
- Self-hosted дешевле для production
- Референс (memos) использует embedded frontend

**План**:
- Спринт 1: Vercel
- Спринт 2: Nginx + embedded frontend (Go embed)

---

## 8. Success Metrics (KPI)

### Спринт 1
- ✅ HTTPS на VPS (Let's Encrypt cert)
- ✅ WebSocket live feed (10+ concurrent connections)
- ✅ Mobile UI polish (responsive на iPhone/Android)
- ✅ CI pipeline (GitHub Actions green)

### Спринт 2
- ✅ RSS feed (50+ subscribers)
- ✅ Export API (10+ exports)
- ✅ Media thumbnails (100% coverage)
- ✅ Custom emoji (10+ sticker packs)

### Спринт 3
- ✅ Микросервисы (5 сервисов running)
- ✅ Kafka events (1000+ events/day)
- ✅ Horizontal scaling (2+ API instances)
- ✅ Monitoring (Grafana dashboards)

---

## 9. Риски и митигации

### Риск 1: Сложность микросервисов
**Митигация**: Начать с monolith, split постепенно.

### Риск 2: Kafka overhead
**Митигация**: Kafka опционален (env var), можно отключить.

### Риск 3: Media storage costs
**Митигация**: MinIO self-hosted (дешевле S3).

### Риск 4: WebSocket scaling
**Митигация**: Redis pub/sub для broadcast между инстансами.

### Риск 5: Security (HTTPS, auth)
**Митигация**: P0 в Спринте 1, Let's Encrypt + JWT rotation.

---

## 10. Следующие шаги (immediate actions)

### Неделя 1 (2026-03-22 → 2026-03-28)
1. [ ] Ротация Telegram/OpenAI токенов
2. [ ] HTTPS на VPS (Let's Encrypt)
3. [ ] golangci-lint setup
4. [ ] air (hot reload) setup
5. [ ] GitHub Actions CI

### Неделя 2 (2026-03-29 → 2026-04-04)
1. [ ] WebSocket endpoint `/api/ws`
2. [ ] Kafka consumer → WS broadcast
3. [ ] Frontend WS subscription
4. [ ] Mobile responsive (feed page)

### Неделя 3 (2026-04-05 → 2026-04-11)
1. [ ] MinIO bucket setup
2. [ ] Media backfill URLs
3. [ ] Admin panel UI (users list)
4. [ ] Sticker packs UI (emoji picker)

### Неделя 4 (2026-04-12 → 2026-04-20)
1. [ ] Unit tests (80% coverage)
2. [ ] Integration tests (WebSocket)
3. [ ] E2E smoke test (Playwright)
4. [ ] Спринт 1 ретроспектива

---

## Заключение

90-дневный план фокусируется на:
1. **Стабилизация** (Спринт 1): HTTPS, WebSocket, mobile UI, CI
2. **Federation** (Спринт 2): RSS, export, media processing, custom emoji
3. **Scale** (Спринт 3): микросервисы, Kafka, horizontal scaling, monitoring

Референсы (memos, bluesky, tinode, GoToSocial) дают паттерны для:
- Минималистичный API (memos)
- Event streams (bluesky)
- Real-time (tinode)
- Media processing (GoToSocial)

Highest-leverage tooling:
- golangci-lint, air, sqlc, testcontainers-go (Спринт 1)
- oapi-codegen, go-swagger (Спринт 2)
- Prometheus, OpenTelemetry, Temporal (Спринт 3)

Критические решения:
- Monolith → Microservices (постепенно)
- REST (не GraphQL)
- JWT → Session (в Спринте 3)
- Postgres (не NoSQL)
- Vercel → Self-hosted (в Спринте 2)

Success metrics: HTTPS, WebSocket, RSS, микросервисы, monitoring.

Риски: сложность микросервисов, Kafka overhead, media costs, WebSocket scaling, security.

Immediate actions: ротация токенов, HTTPS, golangci-lint, air, GitHub Actions.
