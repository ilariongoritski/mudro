# Microservices Concept (MUDRO)

## Цель
Разделить текущий монолитный runtime на независимые сервисы без мгновенного "big bang" переписывания.

## Базовые сервисы (этап 1)
1. `feed-api`
- Отдает `/feed`, `/api/front`, `/api/posts`, `/healthz`
- Читает из Postgres
- Не хранит bot/agent логику

2. `import-vk`
- Пакетный импорт `vk_wall_*.json`
- Пишет в `posts`, `post_reactions`

3. `import-tg`
- Трансформ и загрузка TG (`tgimport`, `tgload`, `tgcommentsimport`)
- Пишет в `posts`, `post_reactions`, `post_comments`

4. `agent-planner`
- Читает TODO
- Создает задачи в `agent_queue`
- Для risky задач ставит `waiting_approval`

5. `agent-worker`
- Забирает из `agent_queue`
- Исполняет только safe task kinds

6. `reporter`
- Формирует дайджесты
- Публикует отчеты в Telegram

7. `telegram-bot` (отдельно от reporter)
- Управляющие команды
- В будущем: API-only, без прямых shell-подпроцессов

## Как перейти без риска
1. Сначала разделить переменные окружения и контуры запуска.
2. Дальше вынести `agent-planner` и `agent-worker` в отдельные deployment units (уже почти готово).
3. Затем развести `telegram-bot` и `reporter` по отдельным зависимостям.
4. После стабилизации — добавить Kafka для событий, не для критичного sync path.

## Kafka: как использовать в MUDRO
Kafka добавлять как event backbone между импортерами, API и воркерами.

Текущее состояние (2026-02-28):
- runtime-публикация task-событий из `cmd/agent` уже включается флагом `KAFKA_ENABLED=true`
- topic по умолчанию: `mudro.agent.tasks.v1`

Топики (предложение):
1. `mudro.posts.v1`
- событие: post upserted
- producer: `import-vk`, `import-tg`
- consumer: индексация/аналитика/кеш

2. `mudro.comments.v1`
- событие: comment upserted
- producer: `tgcommentsimport`
- consumer: агрегаты feed/reaction counters

3. `mudro.agent.tasks.v1`
- событие: task created/approved/rejected/done/failed
- producer: planner/review/worker
- consumer: мониторинг + алерты + daily digest

4. `mudro.notifications.v1`
- событие: ready-to-send notification
- consumer: reporter/sender workers

Принципы:
- ключи партиционирования: `source_post_id` или `task_id`
- idempotency: `event_id` + dedupe в consumer
- schema versioning: `v1` в topic name + `schema_version` в payload
- retries и DLQ для невалидных payload

## RateLimiter: где нужен
1. Входящий HTTP (`feed-api`)
- per-IP limiter на `/api/*` и `/feed`
- цели: защита от burst и дешёвый anti-abuse

Текущее состояние (2026-02-28):
- локальный in-memory token-bucket уже работает
- опционально доступен distributed limiter через Redis (`REDIS_RATE_LIMIT_ENABLED=true`)

2. Исходящие интеграции
- Telegram send limiter (глобальный + per-chat)
- OpenAI limiter (RPM/TPM budget)
- будущие webhook/notification limiter

3. Worker limiter
- ограничение одновременных task execution
- backoff + jitter при ошибках внешних API

Рекомендованные значения на старте:
- API: `20 rps`, burst `40`
- Telegram sender: `10 msg/sec` глобально + `1 msg/sec` на chat
- Agent worker concurrency: `1-2` до появления метрик

## Минимальные метрики
1. `http_requests_total` + `http_429_total`
2. `agent_tasks_total{status=*}`
3. `kafka_publish_total`, `kafka_consume_lag`
4. `telegram_send_total{status=*}`
