# Review Policy

- Любая значимая правка сопровождается проверкой `go test ./...`.
- Для runtime-контура дополнительно проверять `docker compose -f ops/compose/docker-compose.core.yml config`.
- Перед merge проверять отсутствие активных ссылок на `cmd/api|cmd/agent|cmd/bot|cmd/reporter` вне `legacy/old/*`.
