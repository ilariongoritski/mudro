-- ============================================================
-- 007_constraints.sql
-- Add domain CHECK constraints to casino tables,
-- FK on casino_spins.user_id, and missing blackjack index.
--
-- All CHECK constraints use NOT VALID so the migration is safe
-- on a live database: future writes are validated immediately,
-- existing rows are NOT scanned during the migration.
-- Run VALIDATE CONSTRAINT in a separate low-traffic window if
-- full historical validation is needed.
-- ============================================================


-- ── casino_roulette_rounds ────────────────────────────────────────────────────

alter table casino_roulette_rounds
  add constraint casino_roulette_rounds_status_chk
    check (status in ('betting', 'locking', 'spinning', 'result'))
    not valid;

-- winning_color is nullable before the spin; once set it is one of three values.
alter table casino_roulette_rounds
  add constraint casino_roulette_rounds_winning_color_chk
    check (winning_color is null or winning_color in ('red', 'black', 'green'))
    not valid;


-- ── casino_roulette_bets ──────────────────────────────────────────────────────

-- Exhaustive list mirrors roulettePayout() switch in roulette.go.
alter table casino_roulette_bets
  add constraint casino_roulette_bets_bet_type_chk
    check (bet_type in ('straight', 'red', 'black', 'green', 'odd', 'even', 'low', 'high'))
    not valid;

-- 'placed' is the initial value; settleRouletteRound writes lowercase 'win'/'lost'.
alter table casino_roulette_bets
  add constraint casino_roulette_bets_status_chk
    check (status in ('placed', 'win', 'lost'))
    not valid;


-- ── casino_game_activity ──────────────────────────────────────────────────────

-- game_type values come from every call site of insertActivityTx:
-- slots (002 backfill), roulette, plinko, blackjack, bonus (ClaimSubscriptionBonus).
alter table casino_game_activity
  add constraint casino_game_activity_game_type_chk
    check (game_type in ('slots', 'roulette', 'plinko', 'blackjack', 'bonus'))
    not valid;

-- status values per game type:
--   slots    -> WIN | REFUND | LOST  (CASHOUT from live slotStatus())
--   plinko   -> WIN | LOST | CASHOUT
--   roulette -> WIN | LOST
--   blackjack-> PLAYER | DEALER | PUSH | FINISHED
--              (strings.ToUpper(winner) or "FINISHED" when winner is "")
--   bonus    -> CLAIMED
alter table casino_game_activity
  add constraint casino_game_activity_status_chk
    check (status in (
      'WIN', 'LOST', 'REFUND', 'CASHOUT',
      'FINISHED', 'CLAIMED',
      'PLAYER', 'DEALER', 'PUSH'
    ))
    not valid;


-- ── casino_blackjack_games ────────────────────────────────────────────────────

-- BlackjackStatus consts in models.go.
alter table casino_blackjack_games
  add constraint casino_blackjack_games_status_chk
    check (status in ('player_turn', 'dealer_turn', 'resolved'))
    not valid;

-- winner is NULL until the game ends; comment in models.go: 'player', 'dealer', 'push'.
alter table casino_blackjack_games
  add constraint casino_blackjack_games_winner_chk
    check (winner is null or winner in ('player', 'dealer', 'push'))
    not valid;

-- Index for "get game history for user" queries (complement to the partial
-- active-only index already in 006_blackjack.sql).
create index if not exists casino_blackjack_games_user_created_idx
  on casino_blackjack_games (user_id, created_at desc, id desc);


-- ── casino_bonus_claims ───────────────────────────────────────────────────────

-- bonus.go always inserts with status = 'claimed' (hardcoded literal).
alter table casino_bonus_claims
  add constraint casino_bonus_claims_status_chk
    check (status in ('claimed'))
    not valid;

-- Values come from bonusVerificationResult.Status in bonus_telegram.go.
alter table casino_bonus_claims
  add constraint casino_bonus_claims_verification_status_chk
    check (verification_status in (
      'not_configured', 'verification_required',
      'denied', 'unavailable', 'verified'
    ))
    not valid;


-- ── casino_roulette_sessions ──────────────────────────────────────────────────

-- Status is written from round.Status (RoulettePhase) in cacheRouletteSession.
alter table casino_roulette_sessions
  add constraint casino_roulette_sessions_status_chk
    check (status in ('betting', 'locking', 'spinning', 'result'))
    not valid;


-- ── casino_spins - foreign key ────────────────────────────────────────────────

-- 002_live_roulette.sql backfills all casino_spins rows into casino_players,
-- so referential integrity already holds in practice. NOT VALID skips the
-- full-table scan on deployment; run VALIDATE CONSTRAINT in a maintenance
-- window once you want to formally certify the historical data.
alter table casino_spins
  add constraint casino_spins_player_fk
    foreign key (user_id)
    references casino_players(user_id)
    on delete cascade
    not valid;
