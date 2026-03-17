-- MUDRO 010_user_roles_and_subscriptions.sql

-- Добавляем колонку роли в таблицу пользователей
alter table users add column if not exists role text not null default 'user';

-- Создаем таблицу для платных подписок
create table if not exists user_subscriptions (
    id bigserial primary key,
    user_id bigint not null references users(id) on delete cascade,
    plan_id text not null, -- например 'premium_monthly', 'vip'
    status text not null default 'active', -- active, expired, cancelled
    starts_at timestamptz not null default now(),
    expires_at timestamptz not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

-- Индекс для быстрого поиска активных подписок пользователя
create index if not exists user_subscriptions_user_id_idx on user_subscriptions(user_id);
create index if not exists user_subscriptions_expires_at_idx on user_subscriptions(expires_at);
