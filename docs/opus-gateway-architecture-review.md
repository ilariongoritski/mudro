# Opus Gateway Architecture Review

Дата: 2026-03-23

## Статус

Текущая реализация `tools/opus-gateway/` архитектурно близка к рабочему локальному sidecar-сервису, но пока не готова к общему сохранению как "проверенная и цельная" без ещё одной правки по SDK permissions и одной правки по тестовому контуру.

## Ключевые блокеры

1. Ограничения инструментов сейчас потенциально не гарантированы кодом gateway.
   В `src/runner.ts` gateway одновременно передаёт `allowedTools` и рассчитывает на `canUseTool` как на главный барьер для `Read` / `Edit` / `Bash`. По текущей документации Anthropic permission evaluation сначала применяет rules и permission mode, а `canUseTool` вызывается только если решение не было принято раньше. Это создаёт риск, что часть разрешённых инструментов будет auto-approved до пользовательской проверки.

2. SDK-зависимость зафиксирована на старом имени пакета.
   В `tools/opus-gateway/package.json` используется `@anthropic-ai/claude-code`, тогда как актуальный пакет называется `@anthropic-ai/claude-agent-sdk`. Это риск дрейфа API и документации.

3. Тесты завязаны на локальный путь разработчика.
   В `tools/opus-gateway/test/server.test.ts` используется жёсткий `repoRoot` вида `E:\\mudr\\mudro11`, из-за чего тестовый контур не является переносимым и будет ломаться вне этой машины.

## Что уже выглядит правильно

- HTTP sidecar на `127.0.0.1:8788` сам по себе логичен для этого use case.
- Выбор `read-only` / `edit` режимов правильный по направлению.
- Ограничение работы внутри repo root и логирование в `var/log/opus-gateway/` соответствуют задаче.
- Для этого сценария отдельный MCP proxy не обязателен; локальный HTTP gateway достаточно прост и уместен.

## Что нужно сделать перед общим сохранением

1. Перевести пакет и API-вызов на актуальный `Claude Agent SDK`.
2. Пересобрать permission-модель так, чтобы ограничения `Bash` и файловых путей действительно оставались под контролем gateway.
3. Переписать тесты на временную директорию вместо machine-specific path.
4. Только после этого делать полноценный smoke-test с живым `ANTHROPIC_API_KEY`.

## Источники

- https://platform.claude.com/docs/en/agent-sdk/migration-guide
- https://platform.claude.com/docs/en/agent-sdk/permissions
- https://platform.claude.com/docs/en/agent-sdk/typescript
