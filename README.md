# MUDRO

[![CI](https://github.com/ilariongoritski/mudro/actions/workflows/ci.yml/badge.svg)](https://github.com/ilariongoritski/mudro/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

MUDRO — экспериментальный web-first продукт для работы с контентом из VK и Telegram. Репозиторий объединяет веб-интерфейс, Go API, инструменты импорта данных и отдельные сервисы для чата, игровых сценариев и фоновых задач.

**Статус:** pre-MVP. Проект находится в активной разработке; наличие кода и CI не означает, что production-развёртывание уже готово. Перед публичным запуском необходимы настройка окружения, проверка миграций, ротация любых ранее опубликованных секретов и smoke-тесты.

## Возможности

- Лента контента и инструменты импорта/backfill для VK, Telegram и локальных выгрузок.
- Авторизация и пользовательский интерфейс на React.
- Отдельные API для основной ленты и игровых сценариев.
- Чат, административные поверхности и фоновые задачи.
- Экспериментальный planner/worker-контур для автоматизации.
- Контракты API, миграции БД и CI-проверки.

## Технологии

- **Backend:** Go, PostgreSQL.
- **Frontend:** React, TypeScript, Vite, Redux Toolkit / RTK Query.
- **Инфраструктура:** Docker Compose, Redis, Redpanda/Kafka, MinIO (для соответствующих сценариев).
- **Качество:** Go tests, frontend lint/tests/build, проверка контрактов и миграций в GitHub Actions.

## Архитектура

| Компонент | Назначение |
| --- | --- |
| `services/feed-api` | Основной HTTP API: лента и связанные пользовательские сценарии |
| `services/casino` | Отдельный API игровых сценариев и собственный контур БД |
| `services/agent` | Фоновые задачи и planner/worker-автоматизация |
| `services/bot` | Telegram-бот и управляющие сценарии |
| `services/bff-web` | Агрегация API для веб-клиента |
| `services/movie-catalog` | Отдельный каталоговый сервис |
| `tools/importers`, `tools/backfill` | Импорт, восстановление и перенос данных |
| `contracts` | Контракты HTTP API и событий |

Часть сервисов — задел для дальнейшего разделения приложения; не все компоненты следует считать production-ready.

## Локальная разработка

### Требования

- Go версии, указанной в `go.mod`.
- Docker Engine с Docker Compose.
- Node.js и npm для frontend.

### Основной контур

Из корня репозитория сначала создайте локальный файл окружения и замените все `CHANGE_ME` на уникальные значения. Не коммитьте заполненный `.env`:

```bash
cp .env.example .env
```

Для `POSTGRES_PASSWORD`, `CASINO_POSTGRES_PASSWORD`, `CASINO_INTERNAL_SECRET`, `JWT_SECRET` и `MINIO_ROOT_PASSWORD` используйте отдельные случайные значения. Локальные Compose-конфиги намеренно завершаются ошибкой, если обязательные значения не заданы.

Затем запустите основной контур:

```bash
make core-up
make dbcheck-core
make migrate-runtime
make health-runtime
```

Эти команды поднимают локальный контур, проверяют подключение к БД, применяют предусмотренные runtime-миграции и выполняют проверки здоровья. Перед запуском миграций на любой важной БД убедитесь, что используется нужное окружение и есть резервная копия.

### Игровой контур (опционально)

```bash
make casino-up
make health-casino
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Обычно frontend доступен на `http://localhost:5173`, основной API — на `http://localhost:8080`, игровой API — на `http://localhost:8082`. Конкретные порты можно переопределить через переменные окружения.

Для переменных окружения используйте `.env.example` как список доступных настроек и задавайте реальные значения только локально или в менеджере секретов. Не добавляйте реальные токены, пароли, DSN или production-конфигурацию в Git.

## Проверки

Основные команды проекта:

```bash
make selftest
make test-active
go build ./services/... ./tools/... ./cmd/... ./pkg/...
go run ./tools/validate-contracts -dir contracts
```

Frontend:

```bash
cd frontend
npm run lint
npm run test
npm run build
```

Состав тестов и их зависимости зависят от окружения; если проверка требует внешней БД или секретов, это должно быть явно настроено. Не публикуйте вывод конфигурации Compose, если он может содержать значения переменных окружения.

## Docker Compose

- `ops/compose/docker-compose.core.yml` — основной локальный контур.
- `ops/compose/docker-compose.casino.local.yml` — локальный игровой контур.
- `ops/compose/docker-compose.services.yml` — дополнительные сервисы.
- `docker-compose.prod.yml` — production-style конфигурация для самостоятельного развёртывания.
- Корневой `docker-compose.yml` — устаревший совместимый конфиг; для новой разработки используйте файлы из `ops/compose/`.

## Документация

- [Release showcase checklist](docs/release-showcase-checklist.md) — ручная проверка перед демонстрацией.
- [Secrets rotation runbook](ops/runbooks/secrets-rotation.md) — порядок ротации токенов и паролей.
- [VPS/HTTPS runbooks](ops/runbooks/) — материалы по развёртыванию и эксплуатации.
- [CI workflow](.github/workflows/ci.yml) — проверки, выполняемые GitHub Actions.

## Безопасность

- Не храните секреты в репозитории, включая скрипты, примеры конфигурации и тестовые данные.
- Все ранее опубликованные действующие пароли и токены нужно считать скомпрометированными и отозвать/сменить. Удаление из текущей версии не очищает Git history.
- Не используйте демонстрационные учётные данные на публичных или рабочих окружениях.
- Не запускайте production Compose с локальными настройками по умолчанию.

## Лицензия

MIT — см. [LICENSE](./LICENSE).
