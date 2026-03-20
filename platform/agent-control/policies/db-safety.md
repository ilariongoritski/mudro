# DB Safety Policy

- Без явного подтверждения запрещены: `drop`, `truncate`, reset данных, удаление volume.
- Миграции применять только через каноничные команды `make migrate*`.
- Любые изменения схемы, меняющие публичный контракт API, согласовывать отдельно.
- Reporter в legacy-контуре не должен быть частью default production compose.
