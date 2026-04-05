create table if not exists casino_config (
  id boolean primary key default true,
  rtp_percent double precision not null,
  initial_balance bigint not null,
  symbol_weights jsonb not null,
  paytable jsonb not null,
  updated_at timestamptz not null default now()
);

create table if not exists casino_participants (
  user_id bigint primary key,
  username text not null default '',
  email text not null default '',
  role text not null default '',
  coins bigint not null default 500,
  spins_count bigint not null default 0,
  last_spin_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists casino_spins (
  id bigserial primary key,
  user_id bigint not null,
  bet bigint not null,
  win bigint not null,
  symbols jsonb not null,
  game_type text not null default 'slots',
  created_at timestamptz not null default now()
);

create index if not exists casino_spins_user_created_idx
  on casino_spins (user_id, created_at desc, id desc);
