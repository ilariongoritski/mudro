# Codex Profile

- Работать в режиме microservices-first.
- Каноничные runtime entrypoints: `services/feed-api`, `services/agent`, `services/bot`, `services/casino`.
- Для одноразовых CLI использовать `tools/*`; root `cmd/mudro` считать каноническим агрегирующим CLI.
- Не восстанавливать удаленные legacy entrypoints без явной задачи владельца.
- Перед изменениями проверять `git status --short`.
