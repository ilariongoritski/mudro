-- MUDRO 009_agent_events_fk.sql
-- Add missing FK from agent_task_events.task_id -> agent_queue.id

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_name = 'agent_task_events_task_id_fk'
  ) THEN
    ALTER TABLE agent_task_events
      ADD CONSTRAINT agent_task_events_task_id_fk
      FOREIGN KEY (task_id) REFERENCES agent_queue(id) ON DELETE CASCADE;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS agent_task_events_task_id_idx
  ON agent_task_events (task_id);
