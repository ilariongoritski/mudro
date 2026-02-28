-- MUDRO 005_agent_task_events.sql

create table if not exists agent_task_events (
  id bigserial primary key,
  event_id text not null unique,
  task_id bigint not null,
  event_type text not null,
  status text not null,
  kind text,
  dedupe_key text,
  error text,
  occurred_at timestamptz not null default now(),
  created_at timestamptz not null default now()
);

create index if not exists agent_task_events_occurred_idx
  on agent_task_events (occurred_at desc);

create index if not exists agent_task_events_type_occurred_idx
  on agent_task_events (event_type, occurred_at desc);

create index if not exists agent_task_events_kind_occurred_idx
  on agent_task_events (kind, occurred_at desc);
