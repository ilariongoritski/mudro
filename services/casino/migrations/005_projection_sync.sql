alter table if exists casino_players
  add column if not exists wallet_projection_source text not null default 'microservice_projection',
  add column if not exists wallet_projection_note text not null default '',
  add column if not exists wallet_projection_updated_at timestamptz not null default now(),
  add column if not exists wallet_projection_synced_at timestamptz;

create or replace view casino_participants_v2 as
select
  user_id,
  username,
  email,
  role,
  balance as coins,
  games_played as spins_count,
  last_game_at as last_spin_at,
  created_at,
  updated_at
from casino_players;

create table if not exists casino_balance_sync_queue (
  id bigserial primary key,
  user_id bigint not null references casino_players(user_id) on delete cascade,
  reason text not null default '',
  requested_projection_balance bigint,
  status text not null default 'pending'
    check (status in ('pending', 'processing', 'done', 'failed')),
  attempts integer not null default 0,
  available_at timestamptz not null default now(),
  locked_at timestamptz,
  processed_at timestamptz,
  last_error text not null default '',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint casino_balance_sync_queue_user_uq unique (user_id)
);

create index if not exists casino_balance_sync_queue_status_created_idx
  on casino_balance_sync_queue (status, available_at asc, updated_at asc, id asc);

create index if not exists casino_balance_sync_queue_user_created_idx
  on casino_balance_sync_queue (user_id, created_at desc, id desc);

create table if not exists casino_roulette_sessions (
  id bigserial primary key,
  round_id bigint not null unique references casino_roulette_rounds(id) on delete cascade,
  status text not null,
  payload jsonb not null default '{}'::jsonb,
  round_payload jsonb not null default '{}'::jsonb,
  bets_json jsonb not null default '[]'::jsonb,
  expires_at timestamptz not null default now() + interval '5 minutes',
  updated_at timestamptz not null default now(),
  created_at timestamptz not null default now()
);

create index if not exists casino_roulette_sessions_status_updated_idx
  on casino_roulette_sessions (status, updated_at desc, id desc);

create index if not exists casino_roulette_sessions_expires_idx
  on casino_roulette_sessions (expires_at asc, id asc);

create index if not exists casino_game_activity_user_game_created_idx
  on casino_game_activity (user_id, game_type, created_at desc, id desc);
