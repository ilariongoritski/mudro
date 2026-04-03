-- MUDRO 009_agent_events_fk.sql
-- Добавление отсутствующего FK agent_task_events.task_id -> agent_queue.id.

do $$ begin
  if not exists (
    select 1 from information_schema.table_constraints
    where constraint_name = 'agent_task_events_task_id_fk'
  ) then
    alter table agent_task_events
      add constraint agent_task_events_task_id_fk
      foreign key (task_id) references agent_queue(id) on delete cascade;
  end if;
end $$;

create index if not exists agent_task_events_task_id_idx
  on agent_task_events (task_id);
