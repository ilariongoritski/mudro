# Codex Profile

- Работать в режиме microservices-first.
- Каноничные runtime entrypoints: `services/feed-api`, `services/agent`, `services/bot`.
- Не использовать `legacy/old/*` по умолчанию.
- Для CLI использовать `tools/*`; `cmd/*` считать transitional compatibility слоем.
- Перед изменениями проверять `git status --short`.
