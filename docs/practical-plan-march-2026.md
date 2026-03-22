# MUDRO: Практический план на март 2026

**Дата**: 2026-03-22
**Контекст**: Go + PostgreSQL + React/TS, агрегация VK+Telegram, рост в социальную ленту
**Референсы**: usememos/memos, bluesky-social/indigo, tinode/chat, GoToSocial

---

## 1. Что взять из usememos/memos

### Копировать
- **Минималистичный HTTP API**: stdlib, без фреймворков — у нас уже так
- **Единый бинарник с embedded frontend**: Go 1.16+ embed вместо Vercel (экономия $20/мес)
- **Простая JWT auth**: без OAuth сложности — у нас уже есть
- **Webhook интеграции**: у нас Telegram bot, можно добавить Discord/Slack

### Применить к MUDRO
```go
// cmd/api/main.go — embed frontend
//go:embed frontend/dist
var frontendFS embed.FS

http.Handle("/", http.FileServer(http.FS(frontendFS)))
```

**Выгода**: Один бинарник вместо двух деплоев (API + Vercel), проще CI/CD.

---

## 2. Что взять из bluesky-social/indigo

### Копировать
- **Event streams (firehose)**: у нас Kafka, добавить SSE endpoint `/api/firehose`
- **Blob storage**: MinIO для media (уже есть, нужен backfill)
- **Moderation queue**: `agent_queue` для review-задач — уже есть
- **Lexicon schema**: типизированные события для Kafka

### Применить к MUDRO
```go
// internal/events/schema.go
type PostCreated struct {
    Type      string    `json:"type"` // "posts.created"
    PostID    int64     `json:"post_id"`
    UserID    int64     `json:"user_id"`
    Timestamp time.Time `json:"timestamp"`
}
```

**Kafka topics**:
- `posts.created`, `comments.created`, `reactions.added`
- `tasks.created`, `tasks.done`

**SSE endpoint**:
```go
// GET /api/firehose — Server-Sent Events
func (s *Server) handleFirehose(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    // stream Kafka events → SSE
}
```

**Выгода**: Real-time feed без WebSocket сложности, готовая инфра для микросервисов.

---

## 3. Что взять из tinode/chat

### Копировать
- **WebSocket для real-time**: `/api/ws` для live feed updates
- **Presence API**: `GET /api/users/:id/presence` (online/offline)
- **Push notifications**: через Telegram bot — уже есть
- **Message threading**: `parent_comment_id` — уже есть

### Применить к MUDRO
```go
// internal/api/websocket.go
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, _ := upgrader.Upgrade(w, r, nil)
    // subscribe to Kafka → broadcast to WS clients
}
```

**Presence**:
```sql
-- migrations/014_presence.sql
CREATE TABLE user_presence (
    user_id BIGINT PRIMARY KEY REFERENCES users(id),
    status TEXT NOT NULL, -- 'online', 'away', 'offline'
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Выгода**: Live feed без перезагрузки страницы, видно кто онлайн.

---

## 4. Что взять из GoToSocial

### Копировать
- **RSS/Atom feed**: `/api/feed.rss`, `/api/feed.atom` для внешних читалок
- **Media processing**: thumbnails для `preview_url` (resize 300x300, 600x600)
- **Custom emoji**: UI для sticker packs — таблицы есть, нужен UI
- **Export API**: `GET /api/users/:id/export` (JSON archive)

### Применить к MUDRO
```go
// internal/media/thumbnails.go
func GenerateThumbnail(src string, width int) (string, error) {
    // resize image → upload to MinIO → return URL
}
```

**RSS feed**:
```go
// GET /api/feed.rss
func (s *Server) handleRSS(w http.ResponseWriter, r *http.Request) {
    posts := s.posts.LoadPosts(ctx, 50)
    rss := generateRSS(posts)
    w.Header().Set("Content-Type", "application/rss+xml")
    w.Write([]byte(rss))
}
```

**Выгода**: Интеграция с Feedly/Inoreader, экспорт данных для пользователей.

---

## 5. Чего НЕ брать

### usememos/memos
❌ **SQLite в проде** — у нас Postgres, не откатываться
❌ **Отсутствие микросервисов** — у нас уже agent/reporter отдельно
❌ **Отсутствие реакций/комментариев** — у нас есть, не упрощать

### bluesky-social/indigo
❌ **AT Protocol (DID, PDS, BGS)** — слишком сложно для централизованного backend
❌ **Федерация** — у нас один инстанс, не нужна
❌ **IPLD/CAR форматы** — Postgres + JSON проще

### tinode/chat
❌ **Собственный протокол** — REST + JSON достаточно
❌ **gRPC** — HTTP проще для frontend
❌ **Отсутствие SQL** — Postgres лучше для relational data

### GoToSocial
❌ **ActivityPub** — федерация не нужна
❌ **Mastodon API совместимость** — свой API проще поддерживать

---

## 6. 30/60/90-day roadmap

### 30 дней (2026-03-22 → 2026-04-20): Стабилизация + Real-time

**P0 (критично)**:
- [ ] Ротация Telegram/OpenAI токенов (засвечены в git)
- [ ] HTTPS на VPS (Let's Encrypt + Nginx)
- [ ] MinIO bucket setup + backfill URLs в `media_assets`
- [ ] GitHub PAT для MCP (fine-grained, read-only)

**P1 (важно)**:
- [ ] WebSocket `/api/ws` + SSE `/api/firehose`
- [ ] Kafka consumer → WS broadcast (новые посты/комментарии)
- [ ] Frontend: RTK Query + WS subscription
- [ ] Mobile responsive (feed page, drawer, toolbar)
- [ ] Login/Register dark glassmorphism

**P2 (полезно)**:
- [ ] golangci-lint + air (hot reload)
- [ ] GitHub Actions CI: `go test`, `npm run build`, `docker build`
- [ ] Unit tests для `internal/api`, `internal/posts`

**Deliverables**:
✅ HTTPS на VPS
✅ WebSocket live feed
✅ Mobile UI polish
✅ CI pipeline

---

### 60 дней (2026-04-21 → 2026-05-20): Federation + Export

**P0**:
- [ ] RSS feed: `/api/feed.rss` (последние 50 постов)
- [ ] Atom feed: `/api/feed.atom`
- [ ] Per-user feed: `/api/users/:id/feed.rss`

**P1**:
- [ ] Export API: `GET /api/users/:id/export` (JSON archive)
- [ ] Import API: `POST /api/users/:id/import`
- [ ] Thumbnail generation (resize 300x300, 600x600)
- [ ] Preview URL backfill в `media_assets`

**P2**:
- [ ] Sticker packs UI (upload, manage)
- [ ] Emoji picker в comment form
- [ ] Emoji reactions (вместо text reactions)
- [ ] Admin panel UI (users list, stats, agent queue)

**Deliverables**:
✅ RSS/Atom feed
✅ Export/Import API
✅ Media thumbnails
✅ Custom emoji UI

---

### 90 дней (2026-05-21 → 2026-06-20): Microservices + Scale

**P0**:
- [ ] Выделить `services/feed-api` (HTTP API для постов)
- [ ] Выделить `services/agent-planner` (планировщик задач)
- [ ] Выделить `services/agent-worker` (исполнитель задач)
- [ ] Kafka topics: `posts.created`, `comments.created`, `reactions.added`, `tasks.created`, `tasks.done`

**P1**:
- [ ] Redis для distributed rate limiting (уже есть)
- [ ] Redis для session storage (вместо JWT в cookie)
- [ ] Postgres read replicas (для `/api/posts`)
- [ ] Load balancer (Nginx upstream)

**P2**:
- [ ] Prometheus metrics (`/metrics` endpoint)
- [ ] Grafana dashboards (API latency, DB queries, Kafka lag)
- [ ] Structured logging (JSON logs)
- [ ] OpenTelemetry distributed tracing

**Deliverables**:
✅ Микросервисы (5 сервисов)
✅ Kafka event backbone
✅ Horizontal scaling
✅ Monitoring

---

## 7. Highest-leverage actions (топ-5)

### 1. Embedded frontend (неделя 1)
**Зачем**: Один бинарник вместо двух деплоев, экономия $20/мес на Vercel.
**Как**: `go:embed frontend/dist` + `http.FileServer`.
**Референс**: usememos/memos.

### 2. WebSocket + SSE (неделя 2-3)
**Зачем**: Live feed без перезагрузки страницы.
**Как**: `/api/ws` (WebSocket) + `/api/firehose` (SSE fallback).
**Референс**: tinode (WebSocket), bluesky (firehose).

### 3. RSS/Atom feed (неделя 5-6)
**Зачем**: Интеграция с Feedly/Inoreader, рост аудитории.
**Как**: `/api/feed.rss` (XML, последние 50 постов).
**Референс**: GoToSocial.

### 4. Media thumbnails (неделя 7-8)
**Зачем**: Быстрая загрузка feed, экономия трафика.
**Как**: Resize 300x300, 600x600 → MinIO → `preview_url`.
**Референс**: GoToSocial.

### 5. Kafka event backbone (неделя 9-12)
**Зачем**: Готовая инфра для микросервисов, real-time events.
**Как**: Topics `posts.created`, `comments.created`, `reactions.added`.
**Референс**: bluesky (firehose).

---

## 8. Критические решения

### Monolith → Microservices
**Решение**: Начать с monolith (30-60 дней), split в 90 дней.
**Обоснование**: Все референсы начинали с monolith, микросервисы добавляют сложность.

### REST vs GraphQL
**Решение**: REST (текущий подход).
**Обоснование**: Все референсы используют REST, GraphQL сложнее для mobile clients.

### JWT → Session
**Решение**: JWT (30-60 дней) → Redis session (90 дней).
**Обоснование**: JWT проще для stateless API, session лучше для security (revocation).

### Vercel → Self-hosted
**Решение**: Vercel (30 дней) → Embedded frontend (60 дней).
**Обоснование**: Vercel проще для MVP, self-hosted дешевле для production.

---

## 9. Success metrics

### 30 дней
- ✅ HTTPS на VPS (Let's Encrypt cert)
- ✅ WebSocket live feed (10+ concurrent connections)
- ✅ Mobile UI polish (responsive на iPhone/Android)
- ✅ CI pipeline (GitHub Actions green)

### 60 дней
- ✅ RSS feed (50+ subscribers)
- ✅ Export API (10+ exports)
- ✅ Media thumbnails (100% coverage)
- ✅ Custom emoji (10+ sticker packs)

### 90 дней
- ✅ Микросервисы (5 сервисов running)
- ✅ Kafka events (1000+ events/day)
- ✅ Horizontal scaling (2+ API instances)
- ✅ Monitoring (Grafana dashboards)

---

## 10. Immediate actions (неделя 1)

1. **Ротация токенов** (2 часа)
   ```bash
   # Удалить из git history
   git filter-branch --force --index-filter \
     "git rm --cached --ignore-unmatch .env" HEAD
   # Новые токены в .env.local
   ```

2. **HTTPS на VPS** (4 часа)
   ```bash
   sudo apt install certbot python3-certbot-nginx
   sudo certbot --nginx -d api.mudro.ru
   ```

3. **golangci-lint** (1 час)
   ```bash
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
   golangci-lint run
   ```

4. **air (hot reload)** (1 час)
   ```bash
   go install github.com/cosmtrek/air@latest
   air init
   ```

5. **GitHub Actions CI** (2 часа)
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
   ```

**Итого**: 10 часов работы, критичные P0 задачи закрыты.

---

## Заключение

**Что взять**:
1. **memos**: embedded frontend, минималистичный API
2. **bluesky**: event streams (Kafka), blob storage (MinIO), SSE firehose
3. **tinode**: WebSocket, presence API
4. **GoToSocial**: RSS/Atom, media thumbnails, custom emoji, export API

**Чего НЕ брать**:
- Федерация (AT Protocol, ActivityPub)
- Собственные протоколы (gRPC, custom binary)
- SQLite в проде
- GraphQL

**Roadmap**:
- **30 дней**: HTTPS, WebSocket, mobile UI, CI
- **60 дней**: RSS, export, thumbnails, emoji
- **90 дней**: микросервисы, Kafka, scaling, monitoring

**Highest-leverage**:
1. Embedded frontend (экономия $20/мес)
2. WebSocket + SSE (live feed)
3. RSS/Atom (рост аудитории)
4. Media thumbnails (быстрая загрузка)
5. Kafka events (готовая инфра)

**Immediate actions**: ротация токенов, HTTPS, golangci-lint, air, GitHub Actions (10 часов).
