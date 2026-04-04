create table if not exists casino_players (
  user_id bigint primary key,
  username text not null default '',
  email text not null default '',
  role text not null default '',
  display_name text not null default '',
  avatar_url text not null default '',
  telegram_username text not null default '',
  balance bigint not null default 500,
  total_wagered bigint not null default 0,
  total_won bigint not null default 0,
  games_played bigint not null default 0,
  roulette_rounds_played bigint not null default 0,
  level integer not null default 1,
  xp_progress bigint not null default 0,
  last_game_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists casino_game_activity (
  id bigserial primary key,
  user_id bigint not null references casino_players(user_id) on delete cascade,
  game_type text not null,
  game_ref text not null,
  bet_amount bigint not null default 0,
  payout_amount bigint not null default 0,
  net_result bigint not null default 0,
  status text not null,
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);

create unique index if not exists casino_game_activity_game_ref_uq
  on casino_game_activity (game_type, game_ref);

create index if not exists casino_game_activity_user_created_idx
  on casino_game_activity (user_id, created_at desc, id desc);

create table if not exists casino_roulette_rounds (
  id bigserial primary key,
  status text not null,
  winning_number integer,
  winning_color text,
  display_sequence jsonb not null default '[]'::jsonb,
  result_sequence jsonb not null default '[]'::jsonb,
  betting_opens_at timestamptz not null,
  betting_closes_at timestamptz not null,
  spin_started_at timestamptz,
  resolved_at timestamptz,
  created_at timestamptz not null default now()
);

create index if not exists casino_roulette_rounds_status_created_idx
  on casino_roulette_rounds (status, created_at desc, id desc);

create table if not exists casino_roulette_bets (
  id bigserial primary key,
  round_id bigint not null references casino_roulette_rounds(id) on delete cascade,
  user_id bigint not null references casino_players(user_id) on delete cascade,
  bet_type text not null,
  bet_value text not null,
  stake bigint not null,
  payout_amount bigint not null default 0,
  status text not null default 'placed',
  created_at timestamptz not null default now()
);

create unique index if not exists casino_roulette_bets_round_unique_idx
  on casino_roulette_bets (round_id, user_id, bet_type, bet_value);

create index if not exists casino_roulette_bets_round_created_idx
  on casino_roulette_bets (round_id, created_at asc, id asc);

create index if not exists casino_roulette_bets_user_created_idx
  on casino_roulette_bets (user_id, created_at desc, id desc);

with spin_stats as (
  select
    user_id,
    coalesce(sum(bet), 0) as total_wagered,
    coalesce(sum(win), 0) as total_won,
    count(*) as games_played,
    max(created_at) as last_game_at
  from casino_spins
  group by user_id
)
insert into casino_players (
  user_id,
  username,
  email,
  role,
  display_name,
  avatar_url,
  telegram_username,
  balance,
  total_wagered,
  total_won,
  games_played,
  roulette_rounds_played,
  level,
  xp_progress,
  last_game_at,
  created_at,
  updated_at
)
select
  p.user_id,
  p.username,
  p.email,
  p.role,
  '' as display_name,
  '' as avatar_url,
  '' as telegram_username,
  coalesce(p.coins, 500) as balance,
  coalesce(ss.total_wagered, 0) as total_wagered,
  coalesce(ss.total_won, 0) as total_won,
  coalesce(ss.games_played, p.spins_count, 0) as games_played,
  0 as roulette_rounds_played,
  greatest(1, 1 + (coalesce(ss.total_wagered, 0) / 1000)) as level,
  coalesce(ss.total_wagered, 0) % 1000 as xp_progress,
  coalesce(ss.last_game_at, p.last_spin_at) as last_game_at,
  p.created_at,
  p.updated_at
from casino_participants p
left join spin_stats ss on ss.user_id = p.user_id
on conflict (user_id) do update set
  username = excluded.username,
  email = excluded.email,
  role = excluded.role,
  balance = excluded.balance,
  total_wagered = excluded.total_wagered,
  total_won = excluded.total_won,
  games_played = greatest(casino_players.games_played, excluded.games_played),
  roulette_rounds_played = greatest(casino_players.roulette_rounds_played, excluded.roulette_rounds_played),
  level = greatest(casino_players.level, excluded.level),
  xp_progress = greatest(casino_players.xp_progress, excluded.xp_progress),
  last_game_at = coalesce(excluded.last_game_at, casino_players.last_game_at),
  updated_at = now();

with spin_stats as (
  select
    user_id,
    coalesce(sum(bet), 0) as total_wagered,
    coalesce(sum(win), 0) as total_won,
    count(*) as games_played,
    max(created_at) as last_game_at
  from casino_spins
  group by user_id
)
insert into casino_players (
  user_id,
  username,
  email,
  role,
  display_name,
  avatar_url,
  telegram_username,
  balance,
  total_wagered,
  total_won,
  games_played,
  roulette_rounds_played,
  level,
  xp_progress,
  last_game_at,
  created_at,
  updated_at
)
select
  ss.user_id,
  '' as username,
  '' as email,
  '' as role,
  '' as display_name,
  '' as avatar_url,
  '' as telegram_username,
  500 as balance,
  ss.total_wagered,
  ss.total_won,
  ss.games_played,
  0 as roulette_rounds_played,
  greatest(1, 1 + (ss.total_wagered / 1000)) as level,
  ss.total_wagered % 1000 as xp_progress,
  ss.last_game_at,
  now(),
  now()
from spin_stats ss
left join casino_participants p on p.user_id = ss.user_id
where p.user_id is null
on conflict (user_id) do nothing;

insert into casino_game_activity (
  user_id,
  game_type,
  game_ref,
  bet_amount,
  payout_amount,
  net_result,
  status,
  metadata,
  created_at
)
select
  s.user_id,
  'slots',
  s.id::text,
  s.bet,
  s.win,
  s.win - s.bet,
  case
    when s.win > s.bet then 'WIN'
    when s.win = s.bet and s.win > 0 then 'REFUND'
    else 'LOST'
  end,
  jsonb_build_object(
    'symbols', s.symbols,
    'source', 'casino_spins'
  ),
  s.created_at
from casino_spins s
on conflict (game_type, game_ref) do nothing;
