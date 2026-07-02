-- 012: Daily faucet for casino players
alter table casino_players add column if not exists last_faucet_claim timestamptz;
