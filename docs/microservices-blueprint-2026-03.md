# MUDRO Microservices Blueprint (March 2026)

## 1) Цель
Перевести backend MUDRO из mixed-monolith в устойчивую microservices-first архитектуру без big-bang переписывания и без остановки текущего `main` runtime.

Ограничения:
- совместимость текущего `frontend` и `/api/*`;
- локальный запуск через Docker Compose остается обязательным;
- миграция по волнам с rollback-точками.

## 2) Фактическая база (as-is)
Текущие активные сервисы:
- `services/feed-api`
- `services/agent`
- `services/bot`
- `services/casino`

Текущие архитектурные проблемы:
- большой `internal/api/server.go` и смешение нескольких доменов в одном слое;
- миграционный drift в `migrations/` (перекрывающиеся серии auth/casino);
- смешанные API пути (`/api/*`, `/auth/*`, `/casino/*`) и дубли auth-клиентов в frontend;
- coexistence активного `services/casino/*` и legacy `internal/casino/*`.

## 3) Целевая сервисная карта (to-be)

### P0/P1 (обязательно для перехода)
1. `api-gateway`
- Публичный HTTP вход (`/api/v1/*`), auth middleware, rate limit, CORS.
- Не хранит бизнес-данные.

2. `bff-web`
- Агрегация для web/miniapp (`/api/bff/web/v1/*`).
- Не хранит бизнес-данные.

3. `identity-service`
- Пользователи, роли, сессии, Telegram bootstrap auth.
- Владеет auth-таблицами.

4. `feed-query-service`
- Чтение ленты, постов, комментариев (read-optimized).
- Только query-модель.

5. `ingestion-service`
- Запись/импорт постов и комментариев.
- Источник write-событий feed-домена.

6. `casino-service`
- Полностью отдельный домен и отдельная БД (`mudro_casino`).

7. `agent-service`
- Planner/worker lifecycle, очереди задач, event-публикация.

### P2 (после стабилизации)
8. `engagement-service`
- Лайки/реакции/уведомления в отдельном контуре.

9. `media-service`
- Media assets, pre-signed URLs, обработка и выдача media.

10. `bot-service`
- Telegram control plane как тонкий адаптер к API, без прямой DB-логики.

## 4) Каноничная структура директорий

```text
services/
  api-gateway/
  bff-web/
  identity/
  feed-query/
  ingestion/
  engagement/
  casino/
  agent/
  bot/
  media/

contracts/
  grpc/
    identity/v1/*.proto
    feed/v1/*.proto
    casino/v1/*.proto
    agent/v1/*.proto
  http/
    api-gateway-v1.yaml
    bff-web-v1.yaml
  events/
    identity/v1/*.yaml
    feed/v1/*.yaml
    casino/v1/*.yaml
    agent/v1/*.yaml

pkg/
  outbox/
  eventbus/
  grpcutil/
  httputil/
  testutil/

ops/
  compose/
    docker-compose.local.yml
    docker-compose.infra.yml
    docker-compose.services.yml
```

Правила:
- доменная логика только внутри `services/<service>/internal`;
- `pkg/` только технические библиотеки, без доменной логики;
- `contracts/` — source of truth для межсервисных интерфейсов.

## 5) API стратегия и versioning

Публичные слои:
1. Gateway: `/api/v1/*`
2. BFF: `/api/bff/web/v1/*`

Внутренний слой:
- gRPC/Connect API между сервисами (private network).

Versioning policy:
- breaking changes только через новый `vN` в path/proto package;
- additive changes — в пределах текущего major;
- deprecation через `Deprecation` + `Sunset` заголовки;
- минимум 2 релиза параллельной поддержки старого major.

## 6) Data ownership и event-contract

Принцип:
- один сервис — один owner своих таблиц;
- cross-service SQL join запрещен;
- согласованность между доменами через события.

Минимальный outbox шаблон (на каждую сервисную БД):
- `outbox_events(id, aggregate_type, aggregate_id, event_type, event_version, payload, occurred_at, published_at, dedupe_key)`

Минимальный event envelope:
- `event_id`
- `event_type`
- `event_version`
- `producer`
- `aggregate_id`
- `occurred_at`
- `trace_id`
- `payload`

Базовые topics:
- `mudro.auth.user.v1`
- `mudro.feed.post.v1`
- `mudro.feed.comment.v1`
- `mudro.casino.round.v1`
- `mudro.agent.task.v1`

## 7) Волны миграции

### Wave 0 (1-2 недели): стабилизация фундамента
Выход:
- фиксация единого migration source-of-truth;
- устранение drift в `migrations/`;
- внедрение outbox library + первые event contracts;
- единый runtime compose для local.

Rollback:
- оставить текущие маршруты и текущий SQL путь как fallback.

### Wave 1 (2-3 недели): identity extraction
Выход:
- `identity-service` + Telegram bootstrap auth через него;
- gateway auth middleware переключаем на identity API;
- старый auth flow остается за feature flag.

Rollback:
- флаг `USE_IDENTITY_SERVICE=false`.

### Wave 2 (3-4 недели): feed split (query/write)
Выход:
- `ingestion-service` для write;
- `feed-query-service` для read;
- BFF выдает timeline агрегированно;
- старые `/api/posts` и `/api/front` остаются как compat layer.

Rollback:
- трафик обратно на compat layer в gateway.

### Wave 3 (2-3 недели): agent/bot/casino hardening
Выход:
- `agent` разделен на planner и worker deployment units;
- `bot` работает через API-контракты, без прямой привязки к внутренним таблицам feed;
- casino окончательно изолирован, без зависимости от legacy `internal/casino` entrypoints.

Rollback:
- возврат на текущие бинарные entrypoints и compose профили.

## 8) Что нельзя делать (anti-patterns)
- Делить сервисы, но оставлять shared DB ownership.
- Делать синхронную цепочку из 4-5 сервисов в каждом пользовательском запросе.
- Вводить Kafka в критический sync path до стабилизации outbox/idempotency.
- Мигрировать все домены одновременно.

## 9) Первые 3 PR (точно в работу)

### PR-1: Contracts + migration normalization
Scope:
- нормализация `migrations/*`;
- `contracts/http/api-gateway-v1.yaml`;
- `contracts/http/bff-web-v1.yaml`;
- `contracts/events/*` базовые схемы;
- `pkg/outbox/*`.

Acceptance:
- локально применяются миграции без конфликтов;
- outbox пишет и публикует тестовое событие;
- contracts валидируются в CI.

### PR-2: API unification + BFF bootstrap
Scope:
- ввод `services/api-gateway` (proxy/strangler слой);
- ввод `services/bff-web` с `GET /api/bff/web/v1/timeline`;
- frontend переводится на BFF endpoint без ломки старых routes.

Acceptance:
- старый UI и `/tma/casino` продолжают работать;
- новый BFF endpoint обслуживает timeline;
- можно откатить на старый API одним флагом.

### PR-3: Identity service extraction
Scope:
- `services/identity` + JWT/session/telegram bootstrap;
- `api-gateway` auth middleware переключается на identity service;
- deprecate прямой auth путь из старого монолитного слоя.

Acceptance:
- login/register/me/telegram auth работают через identity;
- контрактные тесты на auth проходят;
- fallback на legacy auth возможен через flag.

## 10) Быстрый operational target (локалка + VPS)
Локально:
- один compose профиль для infra (`postgres-main`, `postgres-casino`, `redis`, `kafka`)
- один compose профиль для services (`api-gateway`, `bff-web`, `feed-api compat`, `casino`, `agent-planner`, `agent-worker`, `bot`)

На VPS:
- публичный только `api-gateway` (и frontend static)
- все доменные сервисы и БД только во внутренней сети
- централизованные логи + trace id сквозь gateway/bff/services.

## 11) Рекомендованные внешние референсы (актуальные)
- Go 1.24 release notes: https://go.dev/doc/go1.24
- gRPC health checks: https://grpc.io/docs/guides/health-checking/
- gRPC retry policy: https://grpc.io/docs/guides/retry/
- OpenTelemetry for Go: https://opentelemetry.io/docs/languages/go/
- Connect for Go: https://connectrpc.com/docs/go/getting-started/
- PostgreSQL logical replication: https://www.postgresql.org/docs/current/logical-replication.html
