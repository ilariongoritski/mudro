# Migration Policy (MUDRO)

## Цель
Зафиксировать безопасный и воспроизводимый процесс перехода к microservices архитектуре без big-bang релиза.

## Принципы
1. Incremental only: изменения внедряются волнами.
2. Backward compatible first: новые контракты добавляются до удаления старых.
3. Feature flags on boundary changes: auth/api-routing/data-read-path.
4. Explicit rollback path для каждой волны.
5. Никаких cross-service SQL зависимостей в целевой архитектуре.

## Политика API/контрактов
- Публичные HTTP API:
  - major changes только через новый path version (`/v1`, `/v2`);
  - additive changes разрешены в пределах major.
- Event contracts:
  - additive-only для текущего major topic;
  - breaking change => новый major topic suffix.
- Deprecation:
  - минимум 90 дней;
  - обязательные `Deprecation` и `Sunset` headers на gateway.

## Политика миграций БД
1. Migration source-of-truth должен быть однозначным.
2. Для production-relevant схем:
   - только idempotent SQL;
   - обязательная проверка на staging.
3. Любое изменение, которое может разрушить данные:
   - только после отдельного approve;
   - с подготовленным rollback SQL.
4. Для межсервисной синхронизации используется outbox pattern, а не cross-DB FK.

## Rollout порядок
1. Контракты и outbox foundation.
2. Gateway/BFF слой как strangler.
3. Identity split.
4. Feed query/write split.
5. Agent/bot/casino hardening и cleanup.

## Минимум валидаций перед merge
1. `go test` по затронутым backend пакетам.
2. Frontend build для изменений UI/API контрактов.
3. Контрактная проверка файлов в `contracts/*`.
4. Smoke-run active runtime (`healthz`, базовые API endpoints).

