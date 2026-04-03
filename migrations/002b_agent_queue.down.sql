-- Down migration for 002b_agent_queue.sql
BEGIN;

DROP INDEX IF EXISTS agent_queue_dedupe_live_uq;
DROP INDEX IF EXISTS agent_queue_status_run_after_idx;
DROP TABLE IF EXISTS agent_queue;

COMMIT;
