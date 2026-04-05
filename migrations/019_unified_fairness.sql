-- 1. Add game_type to casino_rounds
ALTER TABLE casino_rounds ADD COLUMN IF NOT EXISTS game_type TEXT NOT NULL DEFAULT 'slots' CHECK (game_type IN ('slots', 'roulette', 'plinko', 'blackjack'));

-- 2. Add same fairness columns to other specialized tables if we keep them for now
-- Note: Stage 1 will eventually migrate everything to casino_rounds.

ALTER TABLE casino_roulette_rounds ADD COLUMN IF NOT EXISTS server_seed TEXT;
ALTER TABLE casino_roulette_rounds ADD COLUMN IF NOT EXISTS server_seed_hash TEXT;
ALTER TABLE casino_roulette_rounds ADD COLUMN IF NOT EXISTS client_seed TEXT;
ALTER TABLE casino_roulette_rounds ADD COLUMN IF NOT EXISTS nonce INT DEFAULT 0;
ALTER TABLE casino_roulette_rounds ADD COLUMN IF NOT EXISTS round_hash TEXT;

ALTER TABLE casino_blackjack_games ADD COLUMN IF NOT EXISTS server_seed TEXT;
ALTER TABLE casino_blackjack_games ADD COLUMN IF NOT EXISTS server_seed_hash TEXT;
ALTER TABLE casino_blackjack_games ADD COLUMN IF NOT EXISTS client_seed TEXT;
ALTER TABLE casino_blackjack_games ADD COLUMN IF NOT EXISTS nonce INT DEFAULT 0;
ALTER TABLE casino_blackjack_games ADD COLUMN IF NOT EXISTS round_hash TEXT;

-- 3. Add metadata to ledger entries
ALTER TABLE casino_ledger_entries ADD COLUMN IF NOT EXISTS metadata JSONB;
