# ENV Profiles

Файл `.env` в корне подходит для локальной разработки, но для прод-контура лучше разносить переменные по сервисам.

Рекомендуемая структура:
- `env/common.env` — общие переменные (`MUDRO_ROOT`, `MUDRO_ENV`, общие feature-flags)
- `env/api.env` — только API (`API_ADDR`, `DSN`, rate-limit настройки)
- `env/agent.env` — только agent (`DSN`, `MUDRO_ROOT`)
- `env/reporter.env` — только reporter (`REPORT_BOT_TOKEN`, обязательный `REPORT_CHAT_ID`, `REPORT_INTERVAL_MIN`)
- `env/bot.env` — только bot (`TELEGRAM_BOT_TOKEN`, `OPENAI_API_KEY`, `API_BASE_URL`)
- `env/db.env` — только db (`POSTGRES_PASSWORD`, `POSTGRES_PORT`)

Дополнительно:
- `env/api.env` поддерживает distributed rate limiting через Redis (`REDIS_RATE_LIMIT_ENABLED`, `REDIS_ADDR`, ...)
- `env/common.env` и `env/agent.env` содержат Kafka runtime-переменные (`KAFKA_ENABLED`, `KAFKA_BROKERS`, `KAFKA_TOPIC_TASKS`)

Шаблоны лежат рядом в `*.env.example`.

Важно:
- `MUDRO_ENV=development` использовать для локалки; `production`/`staging` — для серверных сервисов.
- `user=postgres` в `DSN` допустим только для локальных dev-hosts (`localhost`, `127.0.0.1`, `db`).

Минимальный workflow:
1. Скопировать шаблоны: `cp env/*.env.example env/` (с переименованием, без `.example`).
2. Заполнить секреты в локальных `env/*.env`.
3. Не коммитить реальные `env/*.env`.
