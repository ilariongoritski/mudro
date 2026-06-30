# Review Policy

- Любая значимая правка сопровождается проверкой `make test-active` или точечным `go test -short` по whitelist-пакетам.
- Для runtime-контура дополнительно проверять `docker compose -f ops/compose/docker-compose.core.yml config`.
- Перед merge проверять отсутствие активных ссылок на удаленные legacy runtime entrypoints.
