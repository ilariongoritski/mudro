-- MUDRO 005_agent_review_gate.sql

alter table if exists agent_queue
  drop constraint if exists agent_queue_status_check;

alter table if exists agent_queue
  add constraint agent_queue_status_check
  check (status in ('queued','waiting_approval','in_progress','done','failed','rejected'));

drop index if exists agent_queue_dedupe_live_uq;

create unique index if not exists agent_queue_dedupe_live_uq
  on agent_queue (dedupe_key)
  where dedupe_key is not null and status in ('queued','waiting_approval','in_progress');
