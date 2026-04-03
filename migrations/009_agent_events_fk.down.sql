-- Down migration for 009_agent_events_fk.sql
BEGIN;

DROP INDEX IF EXISTS agent_task_events_task_id_idx;

ALTER TABLE agent_task_events
  DROP CONSTRAINT IF EXISTS agent_task_events_task_id_fk;

COMMIT;
