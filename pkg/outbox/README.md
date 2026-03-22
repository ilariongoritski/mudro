# Outbox Package (Wave-0 Foundation)

`pkg/outbox` provides a minimal, transport-agnostic skeleton for the transactional outbox pattern.

Current scope:
- canonical event envelope (`Event`);
- storage interface (`Store`);
- transport interface (`Publisher`);
- orchestration service (`Service`) with `Enqueue` and `FlushOnce`.

Not included yet:
- concrete SQL implementation for `Store`;
- concrete Kafka/NATS implementation for `Publisher`;
- scheduling loop and backoff strategy.

Expected adoption path:
1. Add `outbox_events` table per service DB.
2. Implement service-local `Store`.
3. Implement shared `Publisher` over chosen event bus.
4. Call `Enqueue` in same transaction as domain state changes.
5. Run dedicated worker calling `FlushOnce`.

Minimal SQL sketch for `outbox_events`:

```sql
create table if not exists outbox_events (
  id bigserial primary key,
  event_type text not null,
  event_version int not null default 1,
  aggregate_id text not null,
  payload jsonb not null,
  trace_id text,
  dedupe_key text,
  occurred_at timestamptz not null default now(),
  published_at timestamptz
);

create index if not exists outbox_events_pending_idx
  on outbox_events (published_at, id)
  where published_at is null;
```

