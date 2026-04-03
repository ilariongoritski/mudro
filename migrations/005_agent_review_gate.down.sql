-- Down migration for 005_agent_review_gate.sql
BEGIN;

-- Revert dedupe index to original (without waiting_approval)
DROP INDEX IF EXISTS agent_queue_dedupe_live_uq;

CREATE UNIQUE INDEX IF NOT EXISTS agent_queue_dedupe_live_uq
  ON agent_queue (dedupe_key)
  WHERE dedupe_key IS NOT NULL AND status IN ('queued','in_progress');

-- Revert status check to original values
ALTER TABLE IF EXISTS agent_queue
  DROP CONSTRAINT IF EXISTS agent_queue_status_check;

ALTER TABLE IF EXISTS agent_queue
  ADD CONSTRAINT agent_queue_status_check
  CHECK (status IN ('queued','in_progress','done','failed'));

COMMIT;
