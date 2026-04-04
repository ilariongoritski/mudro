create table if not exists casino_accounts_audit (
    id bigserial primary key,
    account_id uuid not null references casino_accounts(id) on delete cascade,
    old_balance numeric(30, 10),
    new_balance numeric(30, 10),
    change_amount numeric(30, 10) generated always as (coalesce(new_balance, 0) - coalesce(old_balance, 0)) stored,
    reason text,
    changed_by text not null default 'system',
    changed_at timestamptz not null default now()
);

create index if not exists idx_casino_accounts_audit_account_changed
    on casino_accounts_audit (account_id, changed_at desc);

create or replace function audit_casino_accounts()
returns trigger
language plpgsql
as $$
begin
    if new.balance is distinct from old.balance then
        insert into casino_accounts_audit (
            account_id,
            old_balance,
            new_balance,
            reason,
            changed_by
        )
        values (
            new.id,
            old.balance,
            new.balance,
            nullif(current_setting('app.casino_reason', true), ''),
            coalesce(nullif(current_setting('app.casino_changed_by', true), ''), 'system')
        );
    end if;
    return new;
end;
$$;

drop trigger if exists tr_audit_casino_accounts on casino_accounts;
create trigger tr_audit_casino_accounts
after update on casino_accounts
for each row
execute function audit_casino_accounts();

create table if not exists casino_rtp_tiers (
    id bigserial primary key,
    rtp_profile_id uuid not null references casino_rtp_profiles(id) on delete cascade,
    min_roll integer not null,
    max_roll integer not null,
    multiplier numeric(10, 4) not null,
    label text not null,
    symbol text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    constraint casino_rtp_tiers_range_ck check (min_roll <= max_roll),
    constraint casino_rtp_tiers_profile_range_uq unique (rtp_profile_id, min_roll, max_roll)
);

create index if not exists idx_casino_rtp_tiers_profile_roll
    on casino_rtp_tiers (rtp_profile_id, min_roll, max_roll);

insert into casino_rtp_tiers (
    rtp_profile_id,
    min_roll,
    max_roll,
    multiplier,
    label,
    symbol,
    updated_at
)
select
    p.id,
    (tier.value ->> 'minRoll')::integer,
    (tier.value ->> 'maxRoll')::integer,
    (tier.value ->> 'multiplier')::numeric(10, 4),
    coalesce(tier.value ->> 'label', ''),
    coalesce(tier.value ->> 'symbol', ''),
    now()
from casino_rtp_profiles p
cross join lateral jsonb_array_elements(p.paytable) as tier(value)
on conflict (rtp_profile_id, min_roll, max_roll) do update set
    multiplier = excluded.multiplier,
    label = excluded.label,
    symbol = excluded.symbol,
    updated_at = now();
