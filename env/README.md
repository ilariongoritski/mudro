# ENV Profiles

Файл `.env` в корне подходит для локальной разработки, но для прод-контура лучше разносить переменные по сервисам.

Рекомендуемая структура:
- `env/common.env` — общие переменные (`MUDRO_ROOT`, общие feature-flags)
- `env/api.env` — только API (`API_ADDR`, `DSN`, rate-limit настройки)
- `env/agent.env` — только agent (`DSN`, `MUDRO_ROOT`)
- `env/reporter.env` — только reporter (`REPORT_BOT_TOKEN`, `REPORT_CHAT_ID`, `REPORT_INTERVAL_MIN`)
- `env/bot.env` — только bot (`TELEGRAM_BOT_TOKEN`, `OPENAI_API_KEY`, `API_BASE_URL`)
- `env/db.env` — только db (`POSTGRES_PASSWORD`, `POSTGRES_PORT`)

Шаблоны лежат рядом в `*.env.example`.

Минимальный workflow:
1. Скопировать шаблоны: `cp env/*.env.example env/` (с переименованием, без `.example`).
2. Заполнить секреты в локальных `env/*.env`.
3. Не коммитить реальные `env/*.env`.
