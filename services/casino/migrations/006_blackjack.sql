create table if not exists casino_blackjack_games (
    id bigserial primary key,
    user_id bigint not null references casino_players(user_id) on delete cascade,
    bet bigint not null,
    player_hand jsonb not null default '{}'::jsonb,
    dealer_hand jsonb not null default '{}'::jsonb,
    status text not null,
    winner text, -- 'player', 'dealer', 'push'
    payout bigint not null default 0,
    created_at timestamptz not null default now()
);

create index if not exists casino_blackjack_games_user_status_idx 
    on casino_blackjack_games (user_id, status) where status != 'resolved';
