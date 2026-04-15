alter table if exists casino_blackjack_games
  drop column if exists round_hash,
  drop column if exists nonce,
  drop column if exists client_seed,
  drop column if exists server_seed_hash,
  drop column if exists server_seed;

alter table if exists casino_roulette_rounds
  drop column if exists round_hash,
  drop column if exists nonce,
  drop column if exists client_seed,
  drop column if exists server_seed_hash,
  drop column if exists server_seed;

alter table if exists casino_spins
  drop column if exists nonce,
  drop column if exists client_seed,
  drop column if exists server_seed_hash,
  drop column if exists server_seed;

alter table if exists casino_players
  drop column if exists server_seed,
  drop column if exists server_seed_hash,
  drop column if exists current_nonce,
  drop column if exists client_seed;
