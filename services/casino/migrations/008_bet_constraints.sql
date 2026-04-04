-- Add CHECK constraint for max bet at DB level
-- This provides defense in depth beyond the application-level checks
alter table casino_spins add constraint casino_spins_bet_check 
  check (bet > 0 and bet <= 1000);

alter table casino_roulette_bets add constraint casino_roulette_bets_stake_check 
  check (stake > 0 and stake <= 1000);
