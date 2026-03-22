# TODO (оперативный)

Правила:
- Только конкретные задачи с измеримым результатом.
- Формат: дата, приоритет, зона, цель, следующий шаг.
- Закрытые задачи переносить в `.codex/done.md`.
- Крупные цели хранить в `.codex/todo_big.md`, здесь держать исполняемый short-list.

## P0 (сделать в первую очередь)

- [ ] 2026-03-22 | P0 | area:security | Ротация засвеченных секретов (Telegram/OpenAI)
  - Контекст: токены и ключи публиковались, риск компрометации высокий
  - Прогресс 2026-03-22: секреты доступа удалены из `.codex/server-info.md`, но внешняя ротация ключей ещё не выполнена
  - Следующий шаг: перевыпустить токены/ключи и обновить `.env` на VPS

## P1 (ближайшие 1-3 дня)

- [ ] 2026-03-22 | P1 | area:ops | Синхронизировать VPS-код с Git (pull + rebuild)
  - Контекст: после создания roadmap нужно обновить код на сервере
  - Следующий шаг: `git pull` на VPS + перезапуск Docker-контейнеров

- [ ] 2026-03-22 | P1 | area:deploy | Настроить HTTPS через Let's Encrypt на VPS
  - Контекст: сайт работает по HTTP; нужен домен или subdomain
  - Следующий шаг: Nginx + Certbot, перевести frontend rewrites на HTTPS origin

- [ ] 2026-03-18 | P1 | area:frontend-ui | Mobile Polish для Premium UI
  - Контекст: desktop выглядит хорошо, но мобильная версия не проверена
  - Следующий шаг: проверить отступы, drawer и toolbar на телефоне

- [ ] 2026-03-18 | P1 | area:frontend-ui | Стилизация Login/Register под Premium Dark
  - Контекст: формы логина/регистрации используют старые светлые стили
  - Следующий шаг: перевести на dark glassmorphism

- [ ] 2026-03-18 | P1 | area:agent | Восстановить agent-worker для авто-парсинга
  - Контекст: после починки тестов (P0) воркер должен заработать
  - Следующий шаг: перезапустить `mudro-agent-1` и проверить обработку 41 задачи

- [ ] 2026-02-28 | P1 | area:deploy | Подключить публичное хранилище медиа (MinIO/S3)
  - Контекст: MinIO уже в docker-compose.prod.yml; нужно настроить bucket и backfill URLs
  - Следующий шаг: настроить bucket в MinIO, выполнить backfill URL в `posts.media`

- [ ] 2026-03-17 | P1 | area:vps-web | При появлении домена перевести сайт на HTTPS
  - Контекст: nginx:80 работает, проксирует API/media; не хватает TLS
  - Следующий шаг: завести домен, выпустить сертификат, закрыть порт 8080

## P2 (на неделю)

- [ ] 2026-03-22 | P2 | area:devex | Setup golangci-lint + air (hot reload)
  - Контекст: roadmap рекомендует добавить в Спринте 1
  - Следующий шаг: создать `.golangci.yml` и `.air.toml`, добавить в CI

- [ ] 2026-03-22 | P2 | area:ci | GitHub Actions CI pipeline
  - Контекст: roadmap рекомендует добавить в Неделю 1
  - Следующий шаг: создать `.github/workflows/ci.yml` (go test + golangci-lint + docker build)

- [ ] 2026-03-18 | P2 | area:frontend-ui | Доработать Admin Panel UI
  - Контекст: список пользователей, управление ролями, просмотр очереди агента
  - Следующий шаг: спроектировать экран и реализовать

- [ ] 2026-03-16 | P2 | area:devex | Заполнить GitHub PAT и включить `mudro_github` MCP
  - Контекст: MCP падает с `MUDRO_GITHUB_PAT is not set`
  - Следующий шаг: создать fine-grained PAT и записать в secret-файл

- [ ] 2026-02-25 | P2 | area:bot | Доработать `/feed5` и диагностировать нестабильность
  - Контекст: есть зафиксированная проблема feed5 в памяти
  - Следующий шаг: добавить защиту от пустого API-ответа и тест-кейс

- [ ] 2026-02-25 | P2 | area:repo | Убрать/перенести backup/save артефакты из рабочей структуры
  - Контекст: шум в репозитории ухудшает обзор
  - Следующий шаг: инвентаризировать артефакты и вынести в `docs/legacy`

- [ ] 2026-02-28 | P2 | area:repo | Добавить unit-тесты на ключевые bot/api сценарии
  - Контекст: покрытие тестами недостаточное
  - Следующий шаг: уточнить план и выполнить

## Roadmap Sprint 1 (2026-03-22 → 2026-04-20)

### Неделя 1 (2026-03-22 → 2026-03-28)
- [ ] 2026-03-22 | Sprint1 | area:security | Ротация Telegram/OpenAI токенов
- [ ] 2026-03-22 | Sprint1 | area:ops | HTTPS на VPS (Let's Encrypt)
- [ ] 2026-03-22 | Sprint1 | area:devex | golangci-lint setup
- [ ] 2026-03-22 | Sprint1 | area:devex | air (hot reload) setup
- [ ] 2026-03-22 | Sprint1 | area:ci | GitHub Actions CI

### Неделя 2 (2026-03-29 → 2026-04-04)
- [ ] 2026-03-29 | Sprint1 | area:api | WebSocket endpoint `/api/ws`
- [ ] 2026-03-29 | Sprint1 | area:api | Kafka consumer → WS broadcast
- [ ] 2026-03-29 | Sprint1 | area:frontend | Frontend WS subscription
- [ ] 2026-03-29 | Sprint1 | area:frontend-ui | Mobile responsive (feed page)

### Неделя 3 (2026-04-05 → 2026-04-11)
- [ ] 2026-04-05 | Sprint1 | area:ops | MinIO bucket setup
- [ ] 2026-04-05 | Sprint1 | area:data | Media backfill URLs
- [ ] 2026-04-05 | Sprint1 | area:frontend-ui | Admin panel UI (users list)
- [ ] 2026-04-05 | Sprint1 | area:frontend-ui | Sticker packs UI (emoji picker)

### Неделя 4 (2026-04-12 → 2026-04-20)
- [ ] 2026-04-12 | Sprint1 | area:test | Unit tests (80% coverage)
- [ ] 2026-04-12 | Sprint1 | area:test | Integration tests (WebSocket)
- [ ] 2026-04-12 | Sprint1 | area:test | E2E smoke test (Playwright)
- [ ] 2026-04-20 | Sprint1 | area:process | Спринт 1 ретроспектива

## Ссылка на стратегию
- Масштабные цели и дорожная карта: `.codex/todo_big.md`
- 90-дневный архитектурный план: `docs/architecture-roadmap-90d.md`
