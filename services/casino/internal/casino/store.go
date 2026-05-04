package casino

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const configCacheTTL = 30 * time.Second

type cachedConfig struct {
	cfg       Config
	expiresAt time.Time
}

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrRoundClosed         = errors.New("roulette round is closed")
	ErrNoActiveRound       = errors.New("no active roulette round")
)

const levelStepWagered int64 = 1000

type Store struct {
	pool      *pgxpool.Pool
	mainPool  *pgxpool.Pool
	engine    *Engine
	plinko    *PlinkoEngine
	blackjack *BlackjackEngine

	cfgMu    sync.RWMutex
	cfgCache *cachedConfig
}

func NewStore(pool *pgxpool.Pool, engine *Engine) *Store {
	return NewStoreWithMainPool(pool, nil, engine)
}

func NewStoreWithMainPool(pool, mainPool *pgxpool.Pool, engine *Engine) *Store {
	if engine == nil {
		engine = NewEngine()
	}
	return &Store{
		pool:      pool,
		mainPool:  mainPool,
		engine:    engine,
		plinko:    NewPlinkoEngine(),
		blackjack: NewBlackjackEngine(),
	}
}

func (s *Store) EnsureSeedConfig(ctx context.Context) error {
	_, err := s.GetConfig(ctx)
	return err
}

func (s *Store) Health(ctx context.Context) error {
	if s.pool == nil {
		return errors.New("casino pool is not configured")
	}
	if err := s.pool.Ping(ctx); err != nil {
		return err
	}
	if s.mainPool != nil {
		return s.mainPool.Ping(ctx)
	}
	return nil
}

func (s *Store) GetConfig(ctx context.Context) (Config, error) {
	s.cfgMu.RLock()
	if s.cfgCache != nil && time.Now().Before(s.cfgCache.expiresAt) {
		cfg := s.cfgCache.cfg
		s.cfgMu.RUnlock()
		return cfg, nil
	}
	s.cfgMu.RUnlock()

	cfg, err := s.fetchConfig(ctx)
	if err != nil {
		return Config{}, err
	}

	s.cfgMu.Lock()
	s.cfgCache = &cachedConfig{cfg: cfg, expiresAt: time.Now().Add(configCacheTTL)}
	s.cfgMu.Unlock()

	return cfg, nil
}

func (s *Store) fetchConfig(ctx context.Context) (Config, error) {
	var cfg Config
	var symbolWeights []byte
	var paytable []byte
	err := s.pool.QueryRow(ctx, `
		select rtp_percent, initial_balance, symbol_weights, paytable, updated_at
		from casino_config
		where id = true
	`).Scan(&cfg.RTPPercent, &cfg.InitialBalance, &symbolWeights, &paytable, &cfg.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			cfg = DefaultConfig()
			if err := s.UpdateConfig(ctx, cfg); err != nil {
				return Config{}, err
			}
			return s.fetchConfig(ctx)
		}
		return Config{}, err
	}
	if err := json.Unmarshal(symbolWeights, &cfg.SymbolWeights); err != nil {
		return Config{}, err
	}
	if err := json.Unmarshal(paytable, &cfg.Paytable); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (s *Store) UpdateConfig(ctx context.Context, cfg Config) error {
	if err := ValidateConfig(cfg); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `
		insert into casino_config (id, rtp_percent, initial_balance, symbol_weights, paytable, updated_at)
		values (true, $1, $2, $3::jsonb, $4::jsonb, now())
		on conflict (id) do update set
			rtp_percent = excluded.rtp_percent,
			initial_balance = excluded.initial_balance,
			symbol_weights = excluded.symbol_weights,
			paytable = excluded.paytable,
			updated_at = now()
	`, cfg.RTPPercent, cfg.InitialBalance, marshalJSON(cfg.SymbolWeights), marshalJSON(cfg.Paytable))
	if err != nil {
		return err
	}
	s.cfgMu.Lock()
	s.cfgCache = nil
	s.cfgMu.Unlock()
	return nil
}

func (s *Store) GetBalance(ctx context.Context, actor ParticipantInput) (int64, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return 0, err
	}
	if err := s.ensurePlayer(ctx, nil, actor, cfg); err != nil {
		return 0, err
	}
	balance, _, _, err := s.getWalletState(ctx, s.pool, actor.UserID)
	return balance, err
}

func (s *Store) GetBalanceDetails(ctx context.Context, actor ParticipantInput) (int64, int64, bool, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return 0, 0, false, err
	}
	if err := s.ensurePlayer(ctx, nil, actor, cfg); err != nil {
		return 0, 0, false, err
	}
	return s.getWalletState(ctx, s.pool, actor.UserID)
}

func (s *Store) GetHistory(ctx context.Context, userID int64, limit int) ([]SpinRecord, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.pool.Query(ctx, `
		select id, user_id, bet, win, symbols, created_at
		from casino_spins
		where user_id = $1
		order by created_at desc, id desc
		limit $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]SpinRecord, 0, limit)
	for rows.Next() {
		var (
			item    SpinRecord
			symbols []byte
		)
		if err := rows.Scan(&item.ID, &item.UserID, &item.Bet, &item.Win, &symbols, &item.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(symbols, &item.Symbols); err != nil {
			return nil, err
		}
		history = append(history, item)
	}
	return history, rows.Err()
}

func (s *Store) GetActivity(ctx context.Context, userID int64, limit int) ([]ActivityItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.pool.Query(ctx, `
		select id, game_type, game_ref, bet_amount, payout_amount, net_result, status, metadata::text, created_at
		from casino_game_activity
		where user_id = $1
		order by created_at desc, id desc
		limit $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ActivityItem, 0, limit)
	for rows.Next() {
		var (
			item         ActivityItem
			metadataText string
		)
		if err := rows.Scan(
			&item.ID,
			&item.GameType,
			&item.GameRef,
			&item.BetAmount,
			&item.PayoutAmount,
			&item.NetResult,
			&item.Status,
			&metadataText,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.Metadata = json.RawMessage(metadataText)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetLiveFeed(ctx context.Context, limit int) ([]LiveFeedItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.pool.Query(ctx, `
		select
			a.id,
			a.game_type,
			a.game_ref,
			a.bet_amount,
			a.payout_amount,
			a.net_result,
			a.status,
			a.metadata::text,
			a.created_at,
			p.user_id,
			p.username,
			p.display_name,
			p.avatar_url
		from casino_game_activity a
		join casino_players p on p.user_id = a.user_id
		order by a.created_at desc, a.id desc
		limit $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]LiveFeedItem, 0, limit)
	for rows.Next() {
		var (
			item         LiveFeedItem
			metadataText string
		)
		if err := rows.Scan(
			&item.ID,
			&item.GameType,
			&item.GameRef,
			&item.BetAmount,
			&item.PayoutAmount,
			&item.NetResult,
			&item.Status,
			&metadataText,
			&item.CreatedAt,
			&item.Player.UserID,
			&item.Player.Username,
			&item.Player.DisplayName,
			&item.Player.AvatarURL,
		); err != nil {
			return nil, err
		}
		item.EventType = strings.ToLower(strings.TrimSpace(item.Status))
		if strings.TrimSpace(item.Player.DisplayName) == "" {
			item.Player.DisplayName = item.Player.Username
		}
		item.Metadata = json.RawMessage(metadataText)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetTopWins(ctx context.Context, limit int) ([]LiveFeedItem, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	rows, err := s.pool.Query(ctx, `
		select
			a.id,
			a.game_type,
			a.game_ref,
			a.bet_amount,
			a.payout_amount,
			a.net_result,
			a.status,
			a.metadata::text,
			a.created_at,
			p.user_id,
			p.username,
			p.display_name,
			p.avatar_url
		from casino_game_activity a
		join casino_players p on p.user_id = a.user_id
		where a.net_result > 0
		  and a.created_at >= now() - interval '24 hours'
		order by a.net_result desc, a.created_at desc, a.id desc
		limit $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]LiveFeedItem, 0, limit)
	for rows.Next() {
		var (
			item         LiveFeedItem
			metadataText string
		)
		if err := rows.Scan(
			&item.ID,
			&item.GameType,
			&item.GameRef,
			&item.BetAmount,
			&item.PayoutAmount,
			&item.NetResult,
			&item.Status,
			&metadataText,
			&item.CreatedAt,
			&item.Player.UserID,
			&item.Player.Username,
			&item.Player.DisplayName,
			&item.Player.AvatarURL,
		); err != nil {
			return nil, err
		}
		item.EventType = "top_win"
		if strings.TrimSpace(item.Player.DisplayName) == "" {
			item.Player.DisplayName = item.Player.Username
		}
		item.Metadata = json.RawMessage(metadataText)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetReactions(ctx context.Context, limit int) ([]ReactionFeedItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.pool.Query(ctx, `
		select
			r.activity_id,
			r.emoji,
			count(*)::bigint as reaction_count,
			max(r.updated_at) as latest_at,
			a.game_type,
			a.net_result,
			a.created_at,
			p.user_id,
			p.username,
			p.display_name,
			p.avatar_url
		from casino_activity_reactions r
		join casino_game_activity a on a.id = r.activity_id
		join casino_players p on p.user_id = a.user_id
		group by
			r.activity_id,
			r.emoji,
			a.game_type,
			a.net_result,
			a.created_at,
			p.user_id,
			p.username,
			p.display_name,
			p.avatar_url
		order by latest_at desc
		limit $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ReactionFeedItem, 0, limit)
	for rows.Next() {
		var item ReactionFeedItem
		if err := rows.Scan(
			&item.ActivityID,
			&item.Emoji,
			&item.Count,
			&item.LatestAt,
			&item.GameType,
			&item.NetResult,
			&item.CreatedAt,
			&item.Player.UserID,
			&item.Player.Username,
			&item.Player.DisplayName,
			&item.Player.AvatarURL,
		); err != nil {
			return nil, err
		}
		if strings.TrimSpace(item.Player.DisplayName) == "" {
			item.Player.DisplayName = item.Player.Username
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) AddReaction(ctx context.Context, actor ParticipantInput, req ReactionRequest) (ReactionFeedItem, error) {
	if req.ActivityID <= 0 {
		return ReactionFeedItem{}, fmt.Errorf("activity_id must be positive")
	}
	req.Emoji = strings.TrimSpace(req.Emoji)
	if req.Emoji == "" {
		return ReactionFeedItem{}, fmt.Errorf("emoji is required")
	}
	if len(req.Emoji) > 16 {
		return ReactionFeedItem{}, fmt.Errorf("emoji is too long")
	}

	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return ReactionFeedItem{}, err
	}
	if err := s.ensurePlayer(ctx, nil, actor, cfg); err != nil {
		return ReactionFeedItem{}, err
	}

	_, err = s.pool.Exec(ctx, `
		insert into casino_activity_reactions (activity_id, user_id, emoji, updated_at)
		values ($1, $2, $3, now())
		on conflict (activity_id, user_id) do update set
			emoji = excluded.emoji,
			updated_at = now()
	`, req.ActivityID, actor.UserID, req.Emoji)
	if err != nil {
		return ReactionFeedItem{}, err
	}

	// Targeted aggregate for the just-upserted (activity_id, emoji) pair —
	// avoids loading all 100 reactions and scanning in Go.
	var item ReactionFeedItem
	err = s.pool.QueryRow(ctx, `
		select
			r.activity_id,
			r.emoji,
			count(*)::bigint  as reaction_count,
			max(r.updated_at) as latest_at,
			a.game_type,
			a.net_result,
			a.created_at,
			p.user_id,
			p.username,
			p.display_name,
			p.avatar_url
		from casino_activity_reactions r
		join casino_game_activity a on a.id = r.activity_id
		join casino_players        p on p.user_id = a.user_id
		where r.activity_id = $1 and r.emoji = $2
		group by
			r.activity_id, r.emoji,
			a.game_type, a.net_result, a.created_at,
			p.user_id, p.username, p.display_name, p.avatar_url
	`, req.ActivityID, req.Emoji).Scan(
		&item.ActivityID,
		&item.Emoji,
		&item.Count,
		&item.LatestAt,
		&item.GameType,
		&item.NetResult,
		&item.CreatedAt,
		&item.Player.UserID,
		&item.Player.Username,
		&item.Player.DisplayName,
		&item.Player.AvatarURL,
	)
	if err != nil {
		return ReactionFeedItem{}, err
	}
	if strings.TrimSpace(item.Player.DisplayName) == "" {
		item.Player.DisplayName = item.Player.Username
	}
	return item, nil
}

func (s *Store) GetProfile(ctx context.Context, actor ParticipantInput, limit int) (PlayerProfile, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return PlayerProfile{}, err
	}
	if err := s.ensurePlayer(ctx, nil, actor, cfg); err != nil {
		return PlayerProfile{}, err
	}

	var profile PlayerProfile
	err = s.pool.QueryRow(ctx, `
		select
			user_id,
			username,
			display_name,
			avatar_url,
			balance,
			free_spins_balance,
			bonus_claim_status,
			bonus_claimed_at,
			bonus_verified_at,
			total_wagered,
			total_won,
			games_played,
			roulette_rounds_played,
			level,
			xp_progress,
			last_game_at,
			client_seed,
			current_nonce,
			server_seed_hash
		from casino_players
		where user_id = $1
	`, actor.UserID).Scan(
		&profile.UserID,
		&profile.Username,
		&profile.DisplayName,
		&profile.AvatarURL,
		&profile.Balance,
		&profile.FreeSpinsBalance,
		&profile.BonusClaimStatus,
		&profile.BonusClaimedAt,
		&profile.BonusVerifiedAt,
		&profile.TotalWagered,
		&profile.TotalWon,
		&profile.GamesPlayed,
		&profile.RouletteRoundsPlayed,
		&profile.Level,
		&profile.XPProgress,
		&profile.LastGameAt,
		&profile.ClientSeed,
		&profile.CurrentNonce,
		&profile.ServerSeedHash,
	)
	if err == nil {
		profile.ProgressTarget = levelStepWagered
	}
	if err != nil {
		return PlayerProfile{}, err
	}
	profile.ProgressTarget = levelStepWagered
	if strings.TrimSpace(profile.DisplayName) == "" {
		profile.DisplayName = profile.Username
	}

	activity, err := s.GetActivity(ctx, actor.UserID, limit)
	if err != nil {
		return PlayerProfile{}, err
	}
	profile.RecentActivity = activity
	return profile, nil
}

func (s *Store) Spin(ctx context.Context, actor ParticipantInput, bet int64) (*SpinResult, error) {
	if bet <= 0 {
		return nil, fmt.Errorf("bet must be positive")
	}
	if maxBet := MaxBet(); maxBet > 0 && bet > maxBet {
		return nil, fmt.Errorf("bet exceeds max bet of %d", maxBet)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	cfg, err := s.getConfigTx(ctx, tx)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePlayer(ctx, tx, actor, cfg); err != nil {
		return nil, err
	}

	balance, freeSpins, clientSeed, serverSeed, nonce, _, err := s.getWalletStateForUpdate(ctx, tx, actor.UserID)
	if err != nil {
		return nil, err
	}
	serverSeed, serverSeedHash, err := s.ensurePlayerServerSeedTx(ctx, tx, actor.UserID, serverSeed)
	if err != nil {
		return nil, err
	}

	freeSpinUsed := false
	if freeSpins > 0 {
		freeSpins--
		freeSpinUsed = true
	} else if balance < bet {
		return nil, ErrInsufficientBalance
	}

	// Configure Engine for Provably Fair
	s.engine.EnableFairness(serverSeed, clientSeed, int64(nonce))
	symbols, win, err := s.engine.Spin(cfg, bet)
	s.engine.DisableFairness() // Cleanup
	if err != nil {
		return nil, err
	}

	newBalance := balance + win
	if !freeSpinUsed {
		newBalance -= bet
	}

	if err := s.setWalletStateTx(ctx, tx, actor.UserID, newBalance, &freeSpins); err != nil {
		return nil, err
	}

	// Increment Nonce
	if _, err := tx.Exec(ctx, `update casino_players set current_nonce = current_nonce + 1 where user_id = $1`, actor.UserID); err != nil {
		return nil, err
	}

	var spinID int64
	if err := tx.QueryRow(ctx, `
        insert into casino_spins (user_id, bet, win, symbols, game_type, server_seed, server_seed_hash, client_seed, nonce)
        values ($1, $2, $3, $4::jsonb, 'slots', $5, $6, $7, $8)
        returning id
    `, actor.UserID, bet, win, marshalJSON(symbols), serverSeed, serverSeedHash, clientSeed, nonce).Scan(&spinID); err != nil {
		return nil, err
	}

	// Unified Ledger Transfer
	if err := s.recordTransferTx(ctx, tx, "bet_stake", actor.UserID, bet, map[string]any{"spin_id": spinID}); err != nil {
		return nil, err
	}

	if err := s.updatePlayerStatsTx(ctx, tx, actor.UserID, bet, win, 1, 0); err != nil {
		return nil, err
	}
	netResult := win - bet
	if freeSpinUsed {
		netResult = win
	}
	if err := s.insertActivityTx(ctx, tx, actor.UserID, "slots", fmt.Sprintf("%d", spinID), bet, win, netResult, slotStatus(win, bet), map[string]any{
		"symbols":            symbols,
		"free_spin_used":     freeSpinUsed,
		"free_spins_balance": freeSpins,
		"nonce":              nonce,
		"client_seed":        clientSeed,
		"server_seed_hash":   serverSeedHash,
	}); err != nil {
		return nil, err
	}
	if err := s.enqueueBalanceSyncTx(ctx, tx, actor.UserID, "slots_spin", &newBalance); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true

	return &SpinResult{
		Balance:          newBalance,
		FreeSpinsBalance: freeSpins,
		FreeSpinUsed:     freeSpinUsed,
		Symbols:          symbols,
		Win:              win,
	}, nil
}

func (s *Store) GetCurrentRouletteState(ctx context.Context, userID int64) (RouletteState, error) {
	if cached, ok, err := s.loadRouletteSession(ctx); err == nil && ok {
		if userID > 0 {
			myBets, betsErr := s.getRouletteBetsForRound(ctx, cached.Round.ID, userID)
			if betsErr != nil && !errors.Is(betsErr, pgx.ErrNoRows) {
				return RouletteState{}, betsErr
			}
			cached.MyBets = myBets
		}
		return cached, nil
	} else if err != nil {
		return RouletteState{}, err
	}

	round, err := s.SyncRouletteRound(ctx)
	if err != nil {
		return RouletteState{}, err
	}

	history, err := s.GetRouletteHistory(ctx, 12)
	if err != nil {
		return RouletteState{}, err
	}

	myBets := make([]RouletteBet, 0)
	if userID > 0 {
		myBets, err = s.getRouletteBetsForRound(ctx, round.ID, userID)
		if err != nil {
			return RouletteState{}, err
		}
	}

	secondsLeft := int64(0)
	now := nowUTC()
	switch round.Status {
	case RoulettePhaseBetting:
		secondsLeft = maxInt64(0, int64(round.BettingClosesAt.Sub(now).Seconds()))
	case RoulettePhaseLocking:
		spinAt := round.BettingClosesAt.Add(RouletteLockDuration())
		secondsLeft = maxInt64(0, int64(spinAt.Sub(now).Seconds()))
	case RoulettePhaseSpinning:
		if round.SpinStartedAt != nil {
			resolveAt := round.SpinStartedAt.Add(RouletteSpinDuration())
			secondsLeft = maxInt64(0, int64(resolveAt.Sub(now).Seconds()))
		}
	case RoulettePhaseResult:
		if round.ResolvedAt != nil {
			nextAt := round.ResolvedAt.Add(RouletteResultDuration())
			secondsLeft = maxInt64(0, int64(nextAt.Sub(now).Seconds()))
		}
	}

	_ = s.cacheRouletteSession(ctx, round, history)

	return RouletteState{
		Round:         round,
		Phase:         round.Status,
		ServerTime:    now,
		SecondsLeft:   secondsLeft,
		RecentResults: history,
		MyBets:        myBets,
	}, nil
}

func (s *Store) SyncRouletteRound(ctx context.Context) (RouletteRound, error) {
	round, err := s.GetOrCreateRouletteRound(ctx)
	if err != nil {
		return RouletteRound{}, err
	}
	for i := 0; i < 4; i++ {
		next, err := s.TransitionRouletteRound(ctx, round)
		if err != nil {
			return RouletteRound{}, err
		}
		if next.ID == round.ID && next.Status == round.Status {
			history, historyErr := s.GetRouletteHistory(ctx, 12)
			if historyErr == nil {
				_ = s.cacheRouletteSession(ctx, next, history)
			}
			return next, nil
		}
		round = next
	}
	history, historyErr := s.GetRouletteHistory(ctx, 12)
	if historyErr == nil {
		_ = s.cacheRouletteSession(ctx, round, history)
	}
	return round, nil
}

func (s *Store) PlaceRouletteBets(ctx context.Context, actor ParticipantInput, roundID int64, bets []RouletteBetInput) (*RoulettePlaceBetsResponse, error) {
	if len(bets) == 0 {
		return nil, fmt.Errorf("at least one bet is required")
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	cfg, err := s.getConfigTx(ctx, tx)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePlayer(ctx, tx, actor, cfg); err != nil {
		return nil, err
	}

	if roundID == 0 {
		latest, err := s.getLatestRouletteRoundTx(ctx, tx)
		if err != nil {
			return nil, err
		}
		roundID = latest.ID
	}

	round, err := s.getRouletteRoundTx(ctx, tx, roundID, true)
	if err != nil {
		return nil, err
	}
	if round.Status != RoulettePhaseBetting || nowUTC().After(round.BettingClosesAt) {
		return nil, ErrRoundClosed
	}

	var totalStake int64
	normalized := make([]RouletteBetInput, 0, len(bets))
	for _, bet := range bets {
		normalizedBet, err := normalizeRouletteBet(bet)
		if err != nil {
			return nil, err
		}
		totalStake += normalizedBet.Stake
		normalized = append(normalized, normalizedBet)
	}

	var balance int64
	if err := tx.QueryRow(ctx, `select balance from casino_players where user_id = $1 for update`, actor.UserID).Scan(&balance); err != nil {
		return nil, err
	}
	if balance < totalStake {
		return nil, ErrInsufficientBalance
	}

	newBalance := balance - totalStake
	if err := s.setBalanceTx(ctx, tx, actor.UserID, newBalance); err != nil {
		return nil, err
	}

	placed := make([]RouletteBet, 0, len(normalized))
	for _, bet := range normalized {
		var placedBet RouletteBet
		err := tx.QueryRow(ctx, `
			insert into casino_roulette_bets (round_id, user_id, bet_type, bet_value, stake, payout_amount, status)
			values ($1, $2, $3, $4, $5, 0, 'placed')
			returning id, round_id, user_id, bet_type, bet_value, stake, payout_amount, status, created_at
		`, roundID, actor.UserID, bet.BetType, bet.BetValue, bet.Stake).Scan(
			&placedBet.ID,
			&placedBet.RoundID,
			&placedBet.UserID,
			&placedBet.BetType,
			&placedBet.BetValue,
			&placedBet.Stake,
			&placedBet.PayoutAmount,
			&placedBet.Status,
			&placedBet.CreatedAt,
		)
		if err != nil {
			if isUniqueViolation(err) {
				return nil, fmt.Errorf("duplicate roulette bet for %s:%s", bet.BetType, bet.BetValue)
			}
			return nil, err
		}
		placed = append(placed, placedBet)

		if err := s.recordTransferTx(ctx, tx, "bet_stake", actor.UserID, bet.Stake, map[string]any{"bet_id": placedBet.ID, "round_id": roundID}); err != nil {
			return nil, err
		}
	}

	if err := s.enqueueBalanceSyncTx(ctx, tx, actor.UserID, "roulette_bet", &newBalance); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true

	return &RoulettePlaceBetsResponse{
		RoundID: roundID,
		Balance: newBalance,
		Bets:    placed,
	}, nil
}

func (s *Store) InstantRouletteSpin(ctx context.Context, actor ParticipantInput, bets []RouletteBetInput) (*RouletteInstantSpinResponse, error) {
	if len(bets) == 0 {
		return nil, fmt.Errorf("at least one bet is required")
	}

	var totalStake int64
	normalized := make([]RouletteBetInput, 0, len(bets))
	for _, bet := range bets {
		norm, err := normalizeRouletteBet(bet)
		if err != nil {
			return nil, err
		}
		totalStake += norm.Stake
		normalized = append(normalized, norm)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	cfg, err := s.getConfigTx(ctx, tx)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePlayer(ctx, tx, actor, cfg); err != nil {
		return nil, err
	}

	balance, _, clientSeed, serverSeed, nonce, _, err := s.getWalletStateForUpdate(ctx, tx, actor.UserID)
	if err != nil {
		return nil, err
	}
	serverSeed, serverSeedHash, err := s.ensurePlayerServerSeedTx(ctx, tx, actor.UserID, serverSeed)
	if err != nil {
		return nil, err
	}

	if balance < totalStake {
		return nil, ErrInsufficientBalance
	}

	winningNumber := drawRouletteNumber(NewFairness(serverSeed, clientSeed, int64(nonce)))
	winningColor := rouletteColor(winningNumber)
	displaySequence := buildRouletteDisplaySequence(winningNumber)
	resultSequence := buildRouletteResultSequence(displaySequence, winningNumber)

	now := nowUTC()
	var roundID int64
	err = tx.QueryRow(ctx, `
		insert into casino_roulette_rounds (
			status, winning_number, winning_color, 
			display_sequence, result_sequence, 
			betting_opens_at, betting_closes_at, 
			spin_started_at, resolved_at,
			server_seed, server_seed_hash, client_seed, nonce
		) values (
			'result', $1, $2, $3::jsonb, $4::jsonb, $5, $5, $5, $5, $6, $7, $8, $9
		) returning id
	`, winningNumber, winningColor, marshalJSON(displaySequence), marshalJSON(resultSequence), now, serverSeed, serverSeedHash, clientSeed, nonce).Scan(&roundID)
	if err != nil {
		return nil, err
	}

	var totalPayout int64
	placedBets := make([]RouletteBet, 0, len(normalized))
	for _, b := range normalized {
		payout, won := roulettePayout(RouletteBet{BetType: b.BetType, BetValue: b.BetValue, Stake: b.Stake}, winningNumber, winningColor)
		totalPayout += payout

		status := "lost"
		if won {
			status = "win"
		}

		var pb RouletteBet
		err := tx.QueryRow(ctx, `
			insert into casino_roulette_bets (round_id, user_id, bet_type, bet_value, stake, payout_amount, status)
			values ($1, $2, $3, $4, $5, $6, $7)
			returning id, round_id, user_id, bet_type, bet_value, stake, payout_amount, status, created_at
		`, roundID, actor.UserID, b.BetType, b.BetValue, b.Stake, payout, status).Scan(
			&pb.ID, &pb.RoundID, &pb.UserID, &pb.BetType, &pb.BetValue, &pb.Stake, &pb.PayoutAmount, &pb.Status, &pb.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		placedBets = append(placedBets, pb)

		// Ledger bet recording
		if err := s.recordTransferTx(ctx, tx, "bet_stake", actor.UserID, b.Stake, map[string]any{"bet_id": pb.ID, "game": "roulette_instant"}); err != nil {
			return nil, err
		}
		// Ledger payout recording
		if payout > 0 {
			if err := s.recordTransferTx(ctx, tx, "bet_payout", actor.UserID, payout, map[string]any{"bet_id": pb.ID, "game": "roulette_instant"}); err != nil {
				return nil, err
			}
		}
	}

	newBalance := balance - totalStake + totalPayout
	if err := s.setWalletStateTx(ctx, tx, actor.UserID, newBalance, nil); err != nil {
		return nil, err
	}

	// Increment Nonce
	if _, err := tx.Exec(ctx, `update casino_players set current_nonce = current_nonce + 1 where user_id = $1`, actor.UserID); err != nil {
		return nil, err
	}

	if err := s.updatePlayerStatsTx(ctx, tx, actor.UserID, totalStake, totalPayout, 1, 1); err != nil {
		return nil, err
	}

	netResult := totalPayout - totalStake
	activityStatus := "LOST"
	if totalPayout > totalStake {
		activityStatus = "WIN"
	} else if totalPayout == totalStake {
		activityStatus = "PUSH"
	}

	if err := s.insertActivityTx(ctx, tx, actor.UserID, "roulette_instant", fmt.Sprintf("%d", roundID), totalStake, totalPayout, netResult, activityStatus, map[string]any{
		"winning_number":   winningNumber,
		"winning_color":    winningColor,
		"result_sequence":  resultSequence,
		"client_seed":      clientSeed,
		"nonce":            nonce,
		"server_seed_hash": serverSeedHash,
	}); err != nil {
		return nil, err
	}

	if err := s.enqueueBalanceSyncTx(ctx, tx, actor.UserID, "roulette_instant", &newBalance); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true

	return &RouletteInstantSpinResponse{
		WinningNumber:   winningNumber,
		WinningColor:    winningColor,
		DisplaySequence: displaySequence,
		ResultSequence:  resultSequence,
		PayoutAmount:    totalPayout,
		Balance:         newBalance,
		Bets:            placedBets,
	}, nil
}
func (s *Store) GetRouletteHistory(ctx context.Context, limit int) ([]RouletteHistoryItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.pool.Query(ctx, `
		select id, winning_number, winning_color, resolved_at
		from casino_roulette_rounds
		where status = 'result' and winning_number is not null and resolved_at is not null
		order by resolved_at desc, id desc
		limit $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]RouletteHistoryItem, 0, limit)
	for rows.Next() {
		var item RouletteHistoryItem
		if err := rows.Scan(&item.RoundID, &item.WinningNumber, &item.WinningColor, &item.ResolvedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetOrCreateRouletteRound(ctx context.Context) (RouletteRound, error) {
	round, err := s.getLatestRouletteRound(ctx)
	if err == nil {
		return round, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return RouletteRound{}, err
	}
	return s.createRouletteRound(ctx, nowUTC())
}

func (s *Store) TransitionRouletteRound(ctx context.Context, round RouletteRound) (RouletteRound, error) {
	now := nowUTC()
	switch round.Status {
	case RoulettePhaseBetting:
		if !now.Before(round.BettingClosesAt) {
			return s.updateRouletteRoundStatus(ctx, round.ID, RoulettePhaseLocking, nil, "", nil, nil, nil, nil)
		}
	case RoulettePhaseLocking:
		spinAt := round.BettingClosesAt.Add(RouletteLockDuration())
		if !now.Before(spinAt) {
			fairness := NewFairness(round.ServerSeed, round.ClientSeed, round.Nonce)
			winningNumber := drawRouletteNumber(fairness)
			winningColor := rouletteColor(winningNumber)
			displaySequence := buildRouletteDisplaySequence(winningNumber)
			resultSequence := buildRouletteResultSequence(displaySequence, winningNumber)
			spinStartedAt := now
			return s.updateRouletteRoundStatus(ctx, round.ID, RoulettePhaseSpinning, &winningNumber, winningColor, &spinStartedAt, nil, displaySequence, resultSequence)
		}
	case RoulettePhaseSpinning:
		if round.SpinStartedAt != nil {
			resolveAt := round.SpinStartedAt.Add(RouletteSpinDuration())
			if !now.Before(resolveAt) {
				resolvedAt := now
				updated, err := s.updateRouletteRoundStatus(ctx, round.ID, RoulettePhaseResult, round.WinningNumber, round.WinningColor, round.SpinStartedAt, &resolvedAt, round.DisplaySequence, round.ResultSequence)
				if err != nil {
					return RouletteRound{}, err
				}
				if err := s.settleRouletteRound(ctx, updated); err != nil {
					return RouletteRound{}, err
				}
				return updated, nil
			}
		}
	case RoulettePhaseResult:
		if round.ResolvedAt != nil {
			nextAt := round.ResolvedAt.Add(RouletteResultDuration())
			if !now.Before(nextAt) {
				return s.createRouletteRound(ctx, now)
			}
		}
	}
	return round, nil
}

func (s *Store) GetPlinkoConfig() PlinkoConfig {
	return s.plinko.Config()
}

func (s *Store) GetPlinkoState(ctx context.Context, actor ParticipantInput) (PlinkoState, error) {
	balance, err := s.GetBalance(ctx, actor)
	if err != nil {
		return PlinkoState{}, err
	}
	return PlinkoState{
		Config:  s.plinko.Config(),
		Balance: balance,
	}, nil
}

func (s *Store) DropPlinko(ctx context.Context, actor ParticipantInput, req PlinkoDropRequest) (*PlinkoDropResult, error) {
	if req.Bet <= 0 {
		return nil, fmt.Errorf("bet must be positive")
	}
	if maxBet := MaxBet(); maxBet > 0 && req.Bet > maxBet {
		return nil, fmt.Errorf("bet exceeds max bet of %d", maxBet)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	cfg, err := s.getConfigTx(ctx, tx)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePlayer(ctx, tx, actor, cfg); err != nil {
		return nil, err
	}

	balance, _, clientSeed, serverSeed, nonce, _, err := s.getWalletStateForUpdate(ctx, tx, actor.UserID)
	if err != nil {
		return nil, err
	}
	serverSeed, serverSeedHash, err := s.ensurePlayerServerSeedTx(ctx, tx, actor.UserID, serverSeed)
	if err != nil {
		return nil, err
	}

	if balance < req.Bet {
		return nil, ErrInsufficientBalance
	}

	drop, err := s.plinko.Drop(req.Bet, req.Risk, NewFairness(serverSeed, clientSeed, int64(nonce)))
	if err != nil {
		return nil, err
	}

	newBalance := balance - req.Bet + drop.Payout
	if err := s.setBalanceTx(ctx, tx, actor.UserID, newBalance); err != nil {
		return nil, err
	}

	// Increment Nonce
	if _, err := tx.Exec(ctx, `update casino_players set current_nonce = current_nonce + 1 where user_id = $1`, actor.UserID); err != nil {
		return nil, err
	}

	if err := s.recordTransferTx(ctx, tx, "bet_stake", actor.UserID, req.Bet, map[string]any{"game": "plinko"}); err != nil {
		return nil, err
	}
	if drop.Payout > 0 {
		if err := s.recordTransferTx(ctx, tx, "bet_payout", actor.UserID, drop.Payout, map[string]any{"game": "plinko"}); err != nil {
			return nil, err
		}
	}
	if err := s.updatePlayerStatsTx(ctx, tx, actor.UserID, req.Bet, drop.Payout, 1, 0); err != nil {
		return nil, err
	}

	gameRef := fmt.Sprintf("%d-%d", actor.UserID, drop.CreatedAt.UnixNano())
	if err := s.insertActivityTx(ctx, tx, actor.UserID, "plinko", gameRef, req.Bet, drop.Payout, drop.Payout-req.Bet, drop.Status, map[string]any{
		"risk":             drop.Risk,
		"path":             drop.Path,
		"rows":             drop.Rows,
		"slot_index":       drop.SlotIndex,
		"multiplier":       drop.Multiplier,
		"nonce":            nonce,
		"client_seed":      clientSeed,
		"server_seed_hash": serverSeedHash,
	}); err != nil {
		return nil, err
	}
	if err := s.enqueueBalanceSyncTx(ctx, tx, actor.UserID, "plinko_drop", &newBalance); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true

	drop.Balance = newBalance
	return drop, nil
}

func (s *Store) ensurePlayer(ctx context.Context, tx pgx.Tx, actor ParticipantInput, cfg Config) error {
	if actor.UserID <= 0 {
		return ErrUnauthorized
	}

	initialBalance := cfg.InitialBalance
	projectionSource := "microservice_projection"
	projectionNote := ""
	var projectionSyncedAt any
	if s.mainPool != nil {
		if mainBalance, err := s.getMainWalletBalance(ctx, actor.UserID); err == nil {
			initialBalance = mainBalance
			projectionSource = "main_wallet_bootstrap"
			projectionNote = "main_wallet_bootstrap"
			projectionSyncedAt = nowUTC()
		}
	}

	queryPlayer := `
		insert into casino_players (
			user_id,
			username,
			email,
			role,
			display_name,
			balance,
			level,
			xp_progress,
			wallet_projection_source,
			wallet_projection_note,
			wallet_projection_updated_at,
			wallet_projection_synced_at,
			updated_at,
			client_seed,
			current_nonce,
			server_seed_hash
		)
		values ($1, $2, $3, $4, $5, $6, 1, 0, $7, $8, now(), $9, now(), 'default', 0, $10)
		on conflict (user_id) do update set
			username = excluded.username,
			email = excluded.email,
			role = excluded.role,
			display_name = case when casino_players.display_name = '' then excluded.display_name else casino_players.display_name end,
			balance = case
				when casino_players.wallet_projection_synced_at is null and excluded.wallet_projection_synced_at is not null
					then excluded.balance
				else casino_players.balance
			end,
			wallet_projection_source = case
				when casino_players.wallet_projection_synced_at is null and excluded.wallet_projection_synced_at is not null
					then excluded.wallet_projection_source
				else casino_players.wallet_projection_source
			end,
			wallet_projection_note = case
				when casino_players.wallet_projection_synced_at is null and excluded.wallet_projection_synced_at is not null
					then excluded.wallet_projection_note
				else casino_players.wallet_projection_note
			end,
			wallet_projection_updated_at = now(),
			wallet_projection_synced_at = coalesce(casino_players.wallet_projection_synced_at, excluded.wallet_projection_synced_at),
			updated_at = now()
	`
	seed, _ := GenerateServerSeed()
	seedHash := HashServerSeed(seed)

	args := []any{
		actor.UserID,
		actor.Username,
		actor.Email,
		actor.Role,
		actor.Username,
		initialBalance,
		projectionSource,
		projectionNote,
		projectionSyncedAt,
		seedHash,
	}

	if tx != nil {
		if _, err := tx.Exec(ctx, queryPlayer, args...); err != nil {
			return err
		}
		return s.ensurePlayerLedgerAccountTx(ctx, tx, actor.UserID, initialBalance)
	}

	if _, err := s.pool.Exec(ctx, queryPlayer, args...); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `
		insert into casino_accounts (user_id, type, code, balance, updated_at)
		values ($1, 'user', $2, $3, now())
		on conflict (code) do nothing
	`, actor.UserID, fmt.Sprintf("USER_%d", actor.UserID), initialBalance)
	return err
}

func (s *Store) ensurePlayerLedgerAccountTx(ctx context.Context, tx pgx.Tx, userID int64, balance int64) error {
	_, err := tx.Exec(ctx, `
		insert into casino_accounts (user_id, type, code, balance, updated_at)
		values ($1, 'user', $2, $3, now())
		on conflict (code) do nothing
	`, userID, fmt.Sprintf("USER_%d", userID), balance)
	return err
}

func (s *Store) getConfigTx(ctx context.Context, tx pgx.Tx) (Config, error) {
	var cfg Config
	var symbolWeights []byte
	var paytable []byte
	err := tx.QueryRow(ctx, `
		select rtp_percent, initial_balance, symbol_weights, paytable, updated_at
		from casino_config
		where id = true
		for update
	`).Scan(&cfg.RTPPercent, &cfg.InitialBalance, &symbolWeights, &paytable, &cfg.UpdatedAt)
	if err != nil {
		return Config{}, err
	}
	if err := json.Unmarshal(symbolWeights, &cfg.SymbolWeights); err != nil {
		return Config{}, err
	}
	if err := json.Unmarshal(paytable, &cfg.Paytable); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (s *Store) setBalanceTx(ctx context.Context, tx pgx.Tx, userID int64, balance int64) error {
	return s.setWalletStateTx(ctx, tx, userID, balance, nil)
}

func (s *Store) setWalletStateTx(ctx context.Context, tx pgx.Tx, userID int64, balance int64, freeSpins *int64) error {
	_, err := tx.Exec(ctx, `
		update casino_players
		set balance = $2,
			free_spins_balance = coalesce($3, free_spins_balance),
			wallet_projection_source = 'microservice_projection',
			wallet_projection_note = 'projection_write',
			wallet_projection_updated_at = now(),
			updated_at = now()
		where user_id = $1
	`, userID, balance, nullableInt64(freeSpins))
	return err
}

func (s *Store) getWalletState(ctx context.Context, q queryable, userID int64) (int64, int64, bool, error) {
	var balance, freeSpins int64
	var bonusClaimed bool
	err := q.QueryRow(ctx, `
		select balance, free_spins_balance, (bonus_claimed_at is not null or bonus_claim_status <> '')
		from casino_players
		where user_id = $1
	`, userID).Scan(&balance, &freeSpins, &bonusClaimed)
	return balance, freeSpins, bonusClaimed, err
}

func (s *Store) getWalletStateForUpdate(ctx context.Context, tx pgx.Tx, userID int64) (int64, int64, string, string, int, bool, error) {
	var balance, freeSpins int64
	var bonusClaimed bool
	var clientSeed, serverSeed string
	var nonce int
	err := tx.QueryRow(ctx, `
		select balance, free_spins_balance, client_seed, coalesce(server_seed, ''), current_nonce, (bonus_claimed_at is not null or bonus_claim_status <> '')
		from casino_players
		where user_id = $1
		for update
	`, userID).Scan(&balance, &freeSpins, &clientSeed, &serverSeed, &nonce, &bonusClaimed)
	return balance, freeSpins, clientSeed, serverSeed, nonce, bonusClaimed, err
}

func (s *Store) ensurePlayerServerSeedTx(ctx context.Context, tx pgx.Tx, userID int64, serverSeed string) (string, string, error) {
	if strings.TrimSpace(serverSeed) != "" {
		return serverSeed, HashServerSeed(serverSeed), nil
	}

	serverSeed, err := GenerateServerSeed()
	if err != nil {
		return "", "", err
	}
	serverSeedHash := HashServerSeed(serverSeed)
	_, err = tx.Exec(ctx, `
		update casino_players
		set server_seed = $2,
			server_seed_hash = $3,
			updated_at = now()
		where user_id = $1
	`, userID, serverSeed, serverSeedHash)
	if err != nil {
		return "", "", err
	}
	return serverSeed, serverSeedHash, nil
}

func (s *Store) updatePlayerStatsTx(ctx context.Context, tx pgx.Tx, userID, wagered, won, gamesPlayed, rouletteRoundsPlayed int64) error {
	// Single UPDATE: level and xp_progress are derived in-place from the new
	// total_wagered value (in PostgreSQL SET, RHS column refs use pre-update values,
	// so `total_wagered + $2` is the new total without a second round-trip).
	_, err := tx.Exec(ctx, `
		update casino_players
		set
			total_wagered          = total_wagered + $2,
			total_won              = total_won + $3,
			games_played           = games_played + $4,
			roulette_rounds_played = roulette_rounds_played + $5,
			last_game_at           = now(),
			level                  = (total_wagered + $2) / $6 + 1,
			xp_progress            = (total_wagered + $2) % $6,
			updated_at             = now()
		where user_id = $1
	`, userID, wagered, won, gamesPlayed, rouletteRoundsPlayed, levelStepWagered)
	return err
}

func (s *Store) insertActivityTx(ctx context.Context, tx pgx.Tx, userID int64, gameType, gameRef string, betAmount, payoutAmount, netResult int64, status string, metadata any) error {
	_, err := tx.Exec(ctx, `
		insert into casino_game_activity (
			user_id, game_type, game_ref, bet_amount, payout_amount, net_result, status, metadata
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8::jsonb)
	`, userID, gameType, gameRef, betAmount, payoutAmount, netResult, status, marshalJSON(metadata))
	return err
}

func (s *Store) getLatestRouletteRound(ctx context.Context) (RouletteRound, error) {
	return s.getRouletteRoundQuery(ctx, s.pool)
}

func (s *Store) getLatestRouletteRoundTx(ctx context.Context, tx pgx.Tx) (RouletteRound, error) {
	return s.getRouletteRoundQuery(ctx, tx)
}

func (s *Store) getRouletteRoundQuery(ctx context.Context, q queryable) (RouletteRound, error) {
	return scanRouletteRound(q.QueryRow(ctx, `
		select
			id,
			status,
			winning_number,
			winning_color,
			display_sequence::text,
			result_sequence::text,
			betting_opens_at,
			betting_closes_at,
			spin_started_at,
			resolved_at,
			created_at,
			server_seed,
			server_seed_hash,
			client_seed,
			nonce,
			round_hash
		from casino_roulette_rounds
		order by id desc
		limit 1
	`))
}

func (s *Store) getRouletteRoundTx(ctx context.Context, tx pgx.Tx, roundID int64, forUpdate bool) (RouletteRound, error) {
	query := `
		select
			id,
			status,
			winning_number,
			winning_color,
			display_sequence::text,
			result_sequence::text,
			betting_opens_at,
			betting_closes_at,
			spin_started_at,
			resolved_at,
			created_at,
			server_seed,
			server_seed_hash,
			client_seed,
			nonce,
			round_hash
		from casino_roulette_rounds
		where id = $1
	`
	if forUpdate {
		query += ` for update`
	}
	return scanRouletteRound(tx.QueryRow(ctx, query, roundID))
}

func (s *Store) createRouletteRound(ctx context.Context, now time.Time) (RouletteRound, error) {
	serverSeed, err := GenerateServerSeed()
	if err != nil {
		return RouletteRound{}, err
	}
	serverSeedHash := HashServerSeed(serverSeed)

	return scanRouletteRound(s.pool.QueryRow(ctx, `
		insert into casino_roulette_rounds (
			status,
			display_sequence,
			result_sequence,
			betting_opens_at,
			betting_closes_at,
			server_seed,
			server_seed_hash
		)
		values ('betting', '[]'::jsonb, '[]'::jsonb, $1, $2, $3, $4)
		returning
			id,
			status,
			winning_number,
			winning_color,
			display_sequence::text,
			result_sequence::text,
			betting_opens_at,
			betting_closes_at,
			spin_started_at,
			resolved_at,
			created_at,
			server_seed,
			server_seed_hash,
			client_seed,
			nonce,
			round_hash
	`, now, now.Add(RouletteBettingDuration()), serverSeed, serverSeedHash))
}

func (s *Store) updateRouletteRoundStatus(
	ctx context.Context,
	roundID int64,
	status RoulettePhase,
	winningNumber *int,
	winningColor string,
	spinStartedAt *time.Time,
	resolvedAt *time.Time,
	displaySequence []int,
	resultSequence []int,
) (RouletteRound, error) {
	return scanRouletteRound(s.pool.QueryRow(ctx, `
		update casino_roulette_rounds
		set status = $2,
			winning_number = $3,
			winning_color = $4,
			display_sequence = $5::jsonb,
			result_sequence = $6::jsonb,
			spin_started_at = $7,
			resolved_at = $8
		where id = $1
		returning
			id,
			status,
			winning_number,
			winning_color,
			display_sequence::text,
			result_sequence::text,
			betting_opens_at,
			betting_closes_at,
			spin_started_at,
			resolved_at,
			created_at,
			server_seed,
			server_seed_hash,
			client_seed,
			nonce,
			round_hash
	`, roundID, status, winningNumber, nullableString(winningColor), marshalJSON(displaySequence), marshalJSON(resultSequence), spinStartedAt, resolvedAt))
}

func (s *Store) recordTransferTx(ctx context.Context, tx pgx.Tx, kind string, userID int64, amount int64, meta map[string]any) error {
	if amount == 0 {
		return nil
	}
	var transferID string
	metadata := marshalJSON(meta)
	if err := tx.QueryRow(ctx, `
		insert into casino_transfers (kind, metadata, created_at)
		values ($1, $2, now())
		returning id
	`, kind, metadata).Scan(&transferID); err != nil {
		return err
	}

	var userAccountID string
	if err := tx.QueryRow(ctx, `select id from casino_accounts where code = $1 and type = 'user'`, fmt.Sprintf("USER_%d", userID)).Scan(&userAccountID); err != nil {
		return err
	}
	var houseAccountID string
	if err := tx.QueryRow(ctx, `select id from casino_accounts where code = $1 and type = 'system'`, "SYSTEM_HOUSE_POOL").Scan(&houseAccountID); err != nil {
		return err
	}

	directionUser := "debit"
	directionHouse := "credit"
	if kind == "bet_payout" || kind == "win" {
		directionUser = "credit"
		directionHouse = "debit"
	}

	if _, err := tx.Exec(ctx, `
		insert into casino_ledger_entries (transfer_id, account_id, direction, amount, metadata, created_at)
		values ($1, $2, $3, $4, $5, now())
	`, transferID, userAccountID, directionUser, amount, metadata); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		insert into casino_ledger_entries (transfer_id, account_id, direction, amount, metadata, created_at)
		values ($1, $2, $3, $4, $5, now())
	`, transferID, houseAccountID, directionHouse, amount, metadata); err != nil {
		return err
	}
	return nil
}

// GetRouletteBetsForRound returns all bets placed by a user in the given round.
// Used by the SSE handler to attach per-user bet data to the shared hub state.
func (s *Store) GetRouletteBetsForRound(ctx context.Context, roundID, userID int64) ([]RouletteBet, error) {
	return s.getRouletteBetsForRound(ctx, roundID, userID)
}

func (s *Store) getRouletteBetsForRound(ctx context.Context, roundID, userID int64) ([]RouletteBet, error) {
	rows, err := s.pool.Query(ctx, `
		select id, round_id, user_id, bet_type, bet_value, stake, payout_amount, status, created_at
		from casino_roulette_bets
		where round_id = $1 and user_id = $2
		order by created_at asc, id asc
	`, roundID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]RouletteBet, 0)
	for rows.Next() {
		var item RouletteBet
		if err := rows.Scan(&item.ID, &item.RoundID, &item.UserID, &item.BetType, &item.BetValue, &item.Stake, &item.PayoutAmount, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) settleRouletteRound(ctx context.Context, round RouletteRound) error {
	if round.WinningNumber == nil {
		return fmt.Errorf("roulette round %d has no winning number", round.ID)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	currentRound, err := s.getRouletteRoundTx(ctx, tx, round.ID, true)
	if err != nil {
		return err
	}
	if currentRound.Status != RoulettePhaseResult {
		return fmt.Errorf("roulette round %d is not ready for settlement", round.ID)
	}

	rows, err := tx.Query(ctx, `
		select id, round_id, user_id, bet_type, bet_value, stake, payout_amount, status, created_at
		from casino_roulette_bets
		where round_id = $1 and status = 'placed'
		order by created_at asc, id asc
	`, round.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type playerDelta struct {
		wagered        int64
		won            int64
		rouletteRounds int64
		gamesPlayed    int64
		balanceDelta   int64
	}

	playerDeltas := make(map[int64]*playerDelta)

	for rows.Next() {
		var bet RouletteBet
		if err := rows.Scan(&bet.ID, &bet.RoundID, &bet.UserID, &bet.BetType, &bet.BetValue, &bet.Stake, &bet.PayoutAmount, &bet.Status, &bet.CreatedAt); err != nil {
			return err
		}
		payout, won := roulettePayout(bet, *round.WinningNumber, round.WinningColor)
		status := "LOST"
		if won {
			status = "WIN"
		}

		// Optimization: In a real high-load scenario, we'd batch these up.
		// For now, let's at least keep the payout processing efficient.
		if _, err := tx.Exec(ctx, `
			update casino_roulette_bets
			set payout_amount = $2,
				status = $3
			where id = $1
		`, bet.ID, payout, strings.ToLower(status)); err != nil {
			return err
		}

		if payout > 0 {
			if err := s.recordTransferTx(ctx, tx, "bet_payout", bet.UserID, payout, map[string]any{"bet_id": bet.ID, "round_id": round.ID}); err != nil {
				return err
			}
		}

		delta := playerDeltas[bet.UserID]
		if delta == nil {
			delta = &playerDelta{
				rouletteRounds: 1,
				gamesPlayed:    1,
			}
			playerDeltas[bet.UserID] = delta
		}
		delta.wagered += bet.Stake
		delta.won += payout
		delta.balanceDelta += payout

		if err := s.insertActivityTx(ctx, tx, bet.UserID, "roulette", fmt.Sprintf("%d:%d", round.ID, bet.ID), bet.Stake, payout, payout-bet.Stake, status, map[string]any{
			"round_id":         round.ID,
			"winning_number":   round.WinningNumber,
			"winning_color":    round.WinningColor,
			"bet_type":         bet.BetType,
			"bet_value":        bet.BetValue,
			"result_sequence":  round.ResultSequence,
			"client_seed":      round.ClientSeed,
			"nonce":            round.Nonce,
			"server_seed_hash": round.ServerSeedHash,
		}); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for userID, delta := range playerDeltas {
		var balance int64
		if err := tx.QueryRow(ctx, `select balance from casino_players where user_id = $1 for update`, userID).Scan(&balance); err != nil {
			return err
		}
		if err := s.setBalanceTx(ctx, tx, userID, balance+delta.balanceDelta); err != nil {
			return err
		}
		if err := s.updatePlayerStatsTx(ctx, tx, userID, delta.wagered, delta.won, delta.gamesPlayed, delta.rouletteRounds); err != nil {
			return err
		}
		targetBalance := balance + delta.balanceDelta
		if err := s.enqueueBalanceSyncTx(ctx, tx, userID, "roulette_settlement", &targetBalance); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true

	history, historyErr := s.GetRouletteHistory(ctx, 12)
	if historyErr == nil {
		_ = s.cacheRouletteSession(ctx, round, history)
	}
	return nil
}

type queryable interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func scanRouletteRound(row pgx.Row) (RouletteRound, error) {
	var (
		item              RouletteRound
		displaySequenceTx string
		resultSequenceTx  string
		winningNumber     *int
		winningColor      *string
	)
	if err := row.Scan(
		&item.ID,
		&item.Status,
		&winningNumber,
		&winningColor,
		&displaySequenceTx,
		&resultSequenceTx,
		&item.BettingOpensAt,
		&item.BettingClosesAt,
		&item.SpinStartedAt,
		&item.ResolvedAt,
		&item.CreatedAt,
		&item.ServerSeed,
		&item.ServerSeedHash,
		&item.ClientSeed,
		&item.Nonce,
		&item.RoundHash,
	); err != nil {
		return RouletteRound{}, err
	}
	item.WinningNumber = winningNumber
	if winningColor != nil {
		item.WinningColor = *winningColor
	}
	if err := json.Unmarshal([]byte(displaySequenceTx), &item.DisplaySequence); err != nil {
		return RouletteRound{}, err
	}
	if err := json.Unmarshal([]byte(resultSequenceTx), &item.ResultSequence); err != nil {
		return RouletteRound{}, err
	}
	return item, nil
}

func normalizeRouletteBet(bet RouletteBetInput) (RouletteBetInput, error) {
	bet.BetType = strings.TrimSpace(strings.ToLower(bet.BetType))
	bet.BetValue = strings.TrimSpace(strings.ToLower(bet.BetValue))
	if bet.Stake <= 0 {
		return RouletteBetInput{}, fmt.Errorf("stake must be positive")
	}
	switch bet.BetType {
	case "straight":
		number, err := parseRouletteNumber(bet.BetValue)
		if err != nil {
			return RouletteBetInput{}, err
		}
		bet.BetValue = fmt.Sprintf("%d", number)
	case "red", "black", "green", "odd", "even", "low", "high", "dozen1", "dozen2", "dozen3":
		if bet.BetValue == "" {
			bet.BetValue = bet.BetType
		}
	default:
		return RouletteBetInput{}, fmt.Errorf("unsupported bet type %q", bet.BetType)
	}
	return bet, nil
}

func computeLevel(totalWagered int64) (int64, int64) {
	if totalWagered < 0 {
		totalWagered = 0
	}
	return totalWagered/levelStepWagered + 1, totalWagered % levelStepWagered
}

func slotStatus(win, bet int64) string {
	switch {
	case win > bet:
		return "WIN"
	case win > 0:
		return "CASHOUT"
	default:
		return "LOST"
	}
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func nullableString(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func nullableInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}
