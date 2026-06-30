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

// Configure Engine for Provably Fair

// Cleanup

// Increment Nonce

// Unified Ledger Transfer

// Ledger bet recording

// Ledger payout recording

// Increment Nonce

// Increment Nonce

// Single UPDATE: level and xp_progress are derived in-place from the new
// total_wagered value (in PostgreSQL SET, RHS column refs use pre-update values,
// so `total_wagered + $2` is the new total without a second round-trip).

// GetRouletteBetsForRound returns all bets placed by a user in the given round.
// Used by the SSE handler to attach per-user bet data to the shared hub state.

// Optimization: In a real high-load scenario, we'd batch these up.
// For now, let's at least keep the payout processing efficient.
