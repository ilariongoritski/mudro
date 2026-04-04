create index if not exists casino_game_activity_user_game_created_idx
  on casino_game_activity (user_id, game_type, created_at desc, id desc);
