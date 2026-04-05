-- Missing indexes for MVP performance

-- Balance sync queue: reconcileNextBalanceSync queries WHERE status = 'pending' AND available_at <= now()
CREATE INDEX IF NOT EXISTS idx_casino_balance_sync_queue_pending
  ON casino_balance_sync_queue(status, available_at)
  WHERE status = 'pending';

-- Roulette rounds: getRouletteRoundQuery does ORDER BY id DESC LIMIT 1
CREATE INDEX IF NOT EXISTS idx_casino_roulette_rounds_status_id
  ON casino_roulette_rounds(status, id DESC);

-- Game activity: GetTopWins queries WHERE net_result > 0 AND created_at >= now() - interval '24 hours'
CREATE INDEX IF NOT EXISTS idx_casino_game_activity_top_wins
  ON casino_game_activity(net_result DESC, created_at DESC)
  WHERE net_result > 0;

-- Posts: LoadPosts does ORDER BY published_at, id with optional source filter
CREATE INDEX IF NOT EXISTS idx_posts_source_published_id
  ON posts(source, published_at DESC, id DESC);

-- Post reactions: loadPostReactions queries WHERE post_id = any($1)
CREATE INDEX IF NOT EXISTS idx_post_reactions_post_id
  ON post_reactions(post_id);

-- Comment reactions: loadCommentReactions queries WHERE comment_id = any($1)
CREATE INDEX IF NOT EXISTS idx_comment_reactions_comment_id
  ON comment_reactions(comment_id);
