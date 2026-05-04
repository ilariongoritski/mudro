-- 019_casino_feed_indexes.sql
-- Performance indexes for casino live-feed queries that are polled every 5-10 seconds.

do $$
begin
    if to_regclass('public.casino_game_activity') is not null then
        -- GetLiveFeed: ORDER BY a.created_at DESC, a.id DESC LIMIT N
        create index if not exists idx_casino_game_activity_feed
            on casino_game_activity (created_at desc, id desc);

        -- GetTopWins: WHERE net_result > 0 AND created_at >= now()-24h ORDER BY net_result DESC
        create index if not exists idx_casino_game_activity_top_wins
            on casino_game_activity (net_result desc, created_at desc)
            where net_result > 0;

        -- GetActivity / GetProfile: WHERE user_id = ? ORDER BY created_at DESC, id DESC
        create index if not exists idx_casino_game_activity_user_feed
            on casino_game_activity (user_id, created_at desc, id desc);
    end if;

    if to_regclass('public.casino_activity_reactions') is not null then
        -- GetReactions: ORDER BY max(r.updated_at) DESC, speeds up sort after GROUP BY
        create index if not exists idx_casino_activity_reactions_updated
            on casino_activity_reactions (updated_at desc);
    end if;

    if to_regclass('public.casino_spins') is not null then
        -- GetHistory (slots): WHERE user_id = ? ORDER BY created_at DESC, id DESC
        create index if not exists idx_casino_spins_user_feed
            on casino_spins (user_id, created_at desc, id desc);
    end if;

    if to_regclass('public.casino_blackjack_games') is not null then
        -- BlackjackGetState / BlackjackStart: WHERE user_id = ? AND status != 'resolved'
        create index if not exists idx_casino_blackjack_games_user_status
            on casino_blackjack_games (user_id, status)
            where status != 'resolved';
    end if;
end $$;
