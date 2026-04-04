alter table if exists casino_players
  add column if not exists free_spins_balance bigint not null default 0,
  add column if not exists bonus_claim_status text not null default '',
  add column if not exists bonus_claimed_at timestamptz,
  add column if not exists bonus_verified_at timestamptz;

create table if not exists casino_bonus_claims (
  id bigserial primary key,
  user_id bigint not null references casino_players(user_id) on delete cascade,
  bonus_type text not null,
  status text not null default 'claimed',
  verification_status text not null default 'verification_required',
  verification_message text not null default '',
  free_spins_granted bigint not null default 0,
  telegram_user_id bigint,
  telegram_username text not null default '',
  telegram_channel text not null default '',
  metadata jsonb not null default '{}'::jsonb,
  claimed_at timestamptz not null default now(),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create unique index if not exists casino_bonus_claims_user_bonus_uq
  on casino_bonus_claims (user_id, bonus_type);

create index if not exists casino_bonus_claims_user_created_idx
  on casino_bonus_claims (user_id, created_at desc, id desc);
