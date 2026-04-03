-- Down migration for 006_agent_task_events.sql
BEGIN;

DROP INDEX IF EXISTS agent_task_events_kind_occurred_idx;
DROP INDEX IF EXISTS agent_task_events_type_occurred_idx;
DROP INDEX IF EXISTS agent_task_events_occurred_idx;
DROP TABLE IF EXISTS agent_task_events;

COMMIT;
