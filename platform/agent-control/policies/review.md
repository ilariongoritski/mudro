# Review Policy

- Любая значимая правка сопровождается проверкой `go test ./...`.
- Для runtime-контура дополнительно проверять `docker compose -f ops/compose/docker-compose.core.yml config`.
- Перед merge проверять отсутствие активных ссылок на legacy runtime entrypoints вне `legacy/old/*`.
