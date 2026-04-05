alter table if exists casino_players
  add column if not exists client_seed text not null default 'default',
  add column if not exists current_nonce integer not null default 0,
  add column if not exists server_seed_hash text,
  add column if not exists server_seed text;

alter table if exists casino_spins
  add column if not exists server_seed text,
  add column if not exists server_seed_hash text,
  add column if not exists client_seed text,
  add column if not exists nonce integer not null default 0;

alter table if exists casino_roulette_rounds
  add column if not exists server_seed text,
  add column if not exists server_seed_hash text,
  add column if not exists client_seed text,
  add column if not exists nonce integer not null default 0,
  add column if not exists round_hash text;

alter table if exists casino_blackjack_games
  add column if not exists server_seed text,
  add column if not exists server_seed_hash text,
  add column if not exists client_seed text,
  add column if not exists nonce integer not null default 0,
  add column if not exists round_hash text;
