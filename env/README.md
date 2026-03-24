# ENV Profiles

Файл `.env` в корне подходит для локальной разработки, но для прод-контура лучше разносить переменные по сервисам.

Рекомендуемая структура:
- `env/common.env` — общие переменные (`MUDRO_ROOT`, общие feature-flags)
- `env/api.env` — только API (`API_ADDR`, `DSN`, `JWT_SECRET`, rate-limit настройки, casino proxy URL)
- `env/agent.env` — только agent (`DSN`, `MUDRO_ROOT`)
- `env/reporter.env` — только reporter (`REPORT_BOT_TOKEN`, `REPORT_CHAT_ID`, `REPORT_INTERVAL_MIN`)
- `env/bot.env` — только bot (`TELEGRAM_BOT_TOKEN`, `OPENAI_API_KEY`, `API_BASE_URL`)
- `env/casino.env` — только casino service (`CASINO_ADDR`, `CASINO_DSN`, `CASINO_INITIAL_COINS`, `CASINO_RTP_BPS`, `CASINO_MAX_BET`)
- `env/db.env` — только db (`POSTGRES_PASSWORD`, `POSTGRES_PORT`)
- `env/storage.env` — только object storage / backup (`MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD`)

Дополнительно:
- `env/api.env` поддерживает distributed rate limiting через Redis (`REDIS_RATE_LIMIT_ENABLED`, `REDIS_ADDR`, ...)
- `env/common.env` и `env/agent.env` содержат Kafka runtime-переменные (`KAFKA_ENABLED`, `KAFKA_BROKERS`, `KAFKA_TOPIC_TASKS`)
- `docker-compose.prod.yml` ожидает реальные локальные `env/*.env` и больше не должен опираться на продовые секреты в tracked-файлах

Шаблоны лежат рядом в `*.env.example`.

Минимальный workflow:
1. Скопировать шаблоны: `cp env/*.env.example env/` (с переименованием, без `.example`).
2. Заполнить секреты в локальных `env/*.env`.
3. Не коммитить реальные `env/*.env`.
