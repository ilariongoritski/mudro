-- Add faucet_claim to casino_transfers kind check constraint
alter table casino_transfers drop constraint if exists casino_transfers_kind_check;
alter table casino_transfers add constraint casino_transfers_kind_check
  check (kind in ('bet_stake', 'bet_payout', 'deposit', 'withdrawal', 'adjustment', 'faucet_claim'));

-- Add faucet to casino_game_activity game_type check constraint
alter table casino_game_activity drop constraint if exists casino_game_activity_game_type_chk;
alter table casino_game_activity add constraint casino_game_activity_game_type_chk
  check (game_type in ('slots', 'roulette', 'roulette_instant', 'plinko', 'blackjack', 'bonus', 'faucet'))
  not valid;
