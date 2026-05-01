# Direct Main Stabilization — 2026-05-01

## Контекст
- Владелец явно выбрал одноразовый режим `direct main` для фиксации прогресса за последние две недели.
- Это исключение из обычной политики Mission/BIBLE. Force push, массовый merge всего dirty diff и публикация без проверок запрещены.
- Scope цикла ограничен стабилизацией: lint/test baseline, Vercel routing, безопасные env examples, runtime entrypoint/status и audit-прогресс.

## Что учитывалось
- `docs/audit/20260420-review.md`: большой branch diff нельзя мержить целиком; нужны атомарные изменения.
- `docs/audit/20260422-review.md`: `vercel.json` должен сохранять `/api/:path*` rewrite и SPA fallback.
- Review finding 2026-05-01: frontend lint падал на `react-hooks/set-state-in-effect` в `FeedControls.tsx`.
- Review finding 2026-05-01: общий `go test ./...` захватывал Go-пакеты из `node_modules`.

## Изменения стабилизационного цикла
- `FeedControls.tsx`: убрана синхронизация локального search state через синхронный `setState` внутри effect-body; debounce и reset остаются рабочими.
- `Makefile`: `make test` переведён на активный whitelist через `test-unit`, чтобы не обходить `frontend/node_modules` и `pixel-agents/**/node_modules`.
- `vercel.json`: сохраняется `/api/:path* -> /api/feed` и SPA fallback `/(.*) -> /index.html`.
- `env/*.env.example`: безопасные шаблоны env без секретов можно публиковать как документацию runtime-контура.
- `services/casino/cmd/casino/main.go`: отдельный canonical entrypoint для casino runtime.

## Проверки перед push
- `npm.cmd run lint`
- `npm.cmd run build`
- `go test -short ./services/...`
- `go test -short ./internal/...`
- `go test -short ./cmd/... ./pkg/... ./tools/...`
- `make test`
- `docker compose -f ops/compose/docker-compose.core.yml config`

## Оставшийся backlog
- P0: ротировать ранее засвеченные Telegram/OpenAI секреты вне репозитория.
- Вернуть обычный PR-only workflow после этого исключения.
- Отдельно разобрать большой frontend/UI diff, не смешивая его со стабилизацией.
- Ввести observability baseline: metrics/tracing хотя бы для `feed-api` и `casino`.
- Решить workspace hygiene: `go.work` или другая изоляция external repos/cache от Go tooling.
- Добавить тесты для `services/agent` и `services/bot`.
