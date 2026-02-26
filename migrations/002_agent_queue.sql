-- MUDRO 002_agent_queue.sql

create table if not exists agent_queue (
  id bigserial primary key,
  kind text not null,
  payload jsonb not null default '{}'::jsonb,
  status text not null default 'queued' check (status in ('queued','in_progress','done','failed')),
  priority int not null default 0,
  attempts int not null default 0,
  max_attempts int not null default 3,
  dedupe_key text,
  run_after timestamptz not null default now(),
  locked_by text,
  locked_at timestamptz,
  last_error text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  finished_at timestamptz
);

create index if not exists agent_queue_status_run_after_idx
  on agent_queue (status, run_after, priority desc, id);

create unique index if not exists agent_queue_dedupe_live_uq
  on agent_queue (dedupe_key)
  where dedupe_key is not null and status in ('queued','in_progress');
