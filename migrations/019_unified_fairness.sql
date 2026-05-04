-- 1. Add game_type to casino_rounds
ALTER TABLE casino_rounds ADD COLUMN IF NOT EXISTS game_type TEXT NOT NULL DEFAULT 'slots' CHECK (game_type IN ('slots', 'roulette', 'plinko', 'blackjack'));

-- 2. Add metadata to main wallet ledger entries.
ALTER TABLE casino_ledger_entries ADD COLUMN IF NOT EXISTS metadata JSONB;

-- Casino-service fairness columns live in services/casino/migrations/010_unified_fairness.sql.
