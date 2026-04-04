create table if not exists casino_activity_reactions (
    id bigserial primary key,
    activity_id bigint not null references casino_game_activity(id) on delete cascade,
    user_id bigint not null,
    emoji text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    constraint casino_activity_reactions_one_per_user unique (activity_id, user_id)
);

create index if not exists casino_activity_reactions_activity_idx
    on casino_activity_reactions (activity_id, created_at desc);

create index if not exists casino_activity_reactions_created_idx
    on casino_activity_reactions (created_at desc);
