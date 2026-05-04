create extension if not exists pgcrypto;

create table if not exists casino_accounts (
  id uuid primary key default gen_random_uuid(),
  user_id bigint,
  type text not null default 'user'
    check (type in ('user', 'system')),
  code text unique not null,
  currency text not null default 'MDR',
  balance numeric(30, 10) not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists idx_casino_accounts_user
  on casino_accounts (user_id);

create table if not exists casino_transfers (
  id uuid primary key default gen_random_uuid(),
  kind text not null
    check (kind in ('bet_stake', 'bet_payout', 'deposit', 'withdrawal', 'adjustment')),
  metadata jsonb,
  created_at timestamptz not null default now()
);

create table if not exists casino_ledger_entries (
  id uuid primary key default gen_random_uuid(),
  transfer_id uuid not null references casino_transfers(id) on delete cascade,
  account_id uuid not null references casino_accounts(id) on delete restrict,
  direction text not null
    check (direction in ('debit', 'credit')),
  amount numeric(30, 10) not null,
  metadata jsonb,
  created_at timestamptz not null default now()
);

create index if not exists idx_casino_ledger_transfer
  on casino_ledger_entries (transfer_id);

create index if not exists idx_casino_ledger_account
  on casino_ledger_entries (account_id);

insert into casino_accounts (type, code, currency, balance)
values
  ('system', 'SYSTEM_HOUSE_POOL', 'MDR', 1000000),
  ('system', 'SYSTEM_SETTLEMENT_POOL', 'MDR', 1000000)
on conflict (code) do nothing;
