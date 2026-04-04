package casino

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	defaultBalanceSyncBatch = 32
	balanceSyncRetryDelay   = 15 * time.Second
	rouletteSessionTTL      = 5 * time.Minute
)

type BalanceDrift struct {
	UserID       int64 `json:"user_id"`
	MicroBalance int64 `json:"micro_balance"`
	MainBalance  int64 `json:"main_balance"`
	Diff         int64 `json:"diff"`
}

type rouletteSessionSnapshot struct {
	Round         RouletteRound         `json:"round"`
	Phase         RoulettePhase         `json:"phase"`
	ServerTime    time.Time             `json:"server_time"`
	SecondsLeft   int64                 `json:"seconds_left"`
	RecentResults []RouletteHistoryItem `json:"recent_results"`
}

func (s *Store) StartBalanceReconciler(ctx context.Context, interval time.Duration) {
	if s == nil || s.mainPool == nil || s.pool == nil {
		return
	}
	if interval <= 0 {
		interval = 15 * time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.RunBalanceReconciler(ctx); err != nil && !errors.Is(err, context.Canceled) {
					log.Printf("casino balance reconciler: %v", err)
				}
			}
		}
	}()
}

func (s *Store) RunBalanceReconciler(ctx context.Context) error {
	if s == nil || s.pool == nil || s.mainPool == nil {
		return nil
	}
	for i := 0; i < defaultBalanceSyncBatch; i++ {
		ok, err := s.reconcileNextBalanceSync(ctx)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}
	return nil
}

func (s *Store) StartRouletteSessionJanitor(ctx context.Context, interval time.Duration) {
	if s == nil || s.pool == nil {
		return
	}
	if interval <= 0 {
		interval = 30 * time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.CleanupExpiredRouletteSessions(ctx); err != nil && !errors.Is(err, context.Canceled) {
					log.Printf("casino roulette session janitor: %v", err)
				}
			}
		}
	}()
}

func (s *Store) CleanupRouletteSessions(ctx context.Context) error {
	return s.CleanupExpiredRouletteSessions(ctx)
}

func (s *Store) CleanupExpiredRouletteSessions(ctx context.Context) error {
	if s == nil || s.pool == nil {
		return nil
	}
	_, err := s.pool.Exec(ctx, `
		delete from casino_roulette_sessions
		where expires_at <= now()
	`)
	return err
}

func (s *Store) RebuildPlayerProjection(ctx context.Context, userID int64) error {
	if s == nil || s.pool == nil || s.mainPool == nil || userID <= 0 {
		return nil
	}
	mainBalance, found, err := s.lookupMainWalletBalance(ctx, userID)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	_, err = s.pool.Exec(ctx, `
		insert into casino_players (
			user_id,
			display_name,
			balance,
			level,
			xp_progress,
			wallet_projection_source,
			wallet_projection_note,
			wallet_projection_updated_at,
			wallet_projection_synced_at,
			updated_at
		)
		values ($1, '', $2, 1, 0, 'main_wallet_reconcile', 'rebuild_player_projection', now(), now(), now())
		on conflict (user_id) do update set
			balance = excluded.balance,
			wallet_projection_source = excluded.wallet_projection_source,
			wallet_projection_note = excluded.wallet_projection_note,
			wallet_projection_updated_at = now(),
			wallet_projection_synced_at = now(),
			updated_at = now()
	`, userID, mainBalance)
	return err
}

func (s *Store) enqueueBalanceSyncTx(ctx context.Context, tx pgx.Tx, userID int64, reason string) error {
	if s == nil || s.mainPool == nil || tx == nil || userID <= 0 {
		return nil
	}
	_, err := tx.Exec(ctx, `
		insert into casino_balance_sync_queue (
			user_id,
			reason,
			status,
			attempts,
			last_error,
			available_at,
			processed_at,
			updated_at
		)
		values ($1, $2, 'pending', 0, '', now(), null, now())
		on conflict (user_id) do update set
			reason = excluded.reason,
			status = 'pending',
			last_error = '',
			available_at = now(),
			processed_at = null,
			updated_at = now()
	`, userID, strings.TrimSpace(reason))
	return err
}

func (s *Store) lookupMainWalletBalance(ctx context.Context, userID int64) (int64, bool, error) {
	if s == nil || s.mainPool == nil || userID <= 0 {
		return 0, false, nil
	}
	var balance int64
	err := s.mainPool.QueryRow(ctx, `
		select round(balance)::bigint
		from casino_accounts
		where code = $1 and type = 'user'
	`, fmt.Sprintf("USER_%d", userID)).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return balance, true, nil
}

func (s *Store) getMainWalletBalance(ctx context.Context, userID int64) (int64, error) {
	balance, found, err := s.lookupMainWalletBalance(ctx, userID)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("main wallet USER_%d not found", userID)
	}
	return balance, nil
}

func (s *Store) GetBalanceDrift(ctx context.Context, userID int64) (*BalanceDrift, error) {
	if s == nil || s.pool == nil || s.mainPool == nil || userID <= 0 {
		return nil, nil
	}
	var microBalance int64
	if err := s.pool.QueryRow(ctx, `select balance from casino_players where user_id = $1`, userID).Scan(&microBalance); err != nil {
		return nil, err
	}
	mainBalance, found, err := s.lookupMainWalletBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &BalanceDrift{
		UserID:       userID,
		MicroBalance: microBalance,
		MainBalance:  mainBalance,
		Diff:         mainBalance - microBalance,
	}, nil
}

func (s *Store) reconcileNextBalanceSync(ctx context.Context) (bool, error) {
	if s == nil || s.pool == nil || s.mainPool == nil {
		return false, nil
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var (
		queueID int64
		userID  int64
		reason  string
	)
	err = tx.QueryRow(ctx, `
		select id, user_id, reason
		from casino_balance_sync_queue
		where status = 'pending' and available_at <= now()
		order by available_at asc, updated_at asc, id asc
		limit 1
		for update skip locked
	`).Scan(&queueID, &userID, &reason)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if _, err := tx.Exec(ctx, `
		update casino_balance_sync_queue
		set status = 'processing',
			attempts = attempts + 1,
			updated_at = now(),
			last_error = ''
		where id = $1
	`, queueID); err != nil {
		return false, err
	}

	mainBalance, found, syncErr := s.lookupMainWalletBalance(ctx, userID)
	if syncErr == nil && found {
		if _, syncErr = tx.Exec(ctx, `
			insert into casino_players (
				user_id,
				display_name,
				balance,
				level,
				xp_progress,
				wallet_projection_source,
				wallet_projection_note,
				wallet_projection_updated_at,
				wallet_projection_synced_at,
				updated_at
			)
			values ($1, '', $2, 1, 0, 'main_wallet_reconcile', $3, now(), now(), now())
			on conflict (user_id) do update set
				balance = excluded.balance,
				wallet_projection_source = excluded.wallet_projection_source,
				wallet_projection_note = excluded.wallet_projection_note,
				wallet_projection_updated_at = now(),
				wallet_projection_synced_at = now(),
				updated_at = now()
		`, userID, mainBalance, fallbackWalletReason(reason, "main_wallet_reconcile")); syncErr == nil {
			_, syncErr = tx.Exec(ctx, `
				update casino_balance_sync_queue
				set status = 'done',
					processed_at = now(),
					updated_at = now(),
					last_error = ''
				where id = $1
			`, queueID)
		}
	}
	if syncErr == nil && !found {
		syncErr = fmt.Errorf("main wallet not found for user %d", userID)
	}
	if syncErr != nil {
		if _, err := tx.Exec(ctx, `
			update casino_balance_sync_queue
			set status = 'pending',
				processed_at = null,
				available_at = now() + $2::interval,
				updated_at = now(),
				last_error = $3
			where id = $1
		`, queueID, formatInterval(balanceSyncRetryDelay), syncErr.Error()); err != nil {
			return false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) loadRouletteSession(ctx context.Context) (RouletteState, bool, error) {
	if s == nil || s.pool == nil {
		return RouletteState{}, false, nil
	}
	var payload string
	err := s.pool.QueryRow(ctx, `
		select payload::text
		from casino_roulette_sessions
		where expires_at > now()
		order by updated_at desc, round_id desc
		limit 1
	`).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return RouletteState{}, false, nil
	}
	if err != nil {
		return RouletteState{}, false, err
	}

	var state RouletteState
	if err := json.Unmarshal([]byte(payload), &state); err != nil {
		return RouletteState{}, false, err
	}
	return state, true, nil
}

func (s *Store) cacheRouletteSession(ctx context.Context, round RouletteRound, history []RouletteHistoryItem) error {
	if s == nil || s.pool == nil || round.ID <= 0 {
		return nil
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

	snapshot := rouletteSessionSnapshot{
		Round:         round,
		Phase:         round.Status,
		ServerTime:    now,
		SecondsLeft:   secondsLeft,
		RecentResults: history,
	}

	expiresAt := round.BettingClosesAt.Add(rouletteSessionTTL)
	if round.ResolvedAt != nil {
		expiresAt = round.ResolvedAt.Add(rouletteSessionTTL)
	}

	_, err := s.pool.Exec(ctx, `
		insert into casino_roulette_sessions (
			round_id,
			status,
			payload,
			round_payload,
			bets_json,
			expires_at,
			updated_at
		)
		values ($1, $2, $3::jsonb, $4::jsonb, '[]'::jsonb, $5, now())
		on conflict (round_id) do update set
			status = excluded.status,
			payload = excluded.payload,
			round_payload = excluded.round_payload,
			bets_json = excluded.bets_json,
			expires_at = excluded.expires_at,
			updated_at = now()
	`, round.ID, round.Status, marshalJSON(snapshot), marshalJSON(round), expiresAt)
	return err
}

func formatInterval(d time.Duration) string {
	return fmt.Sprintf("%f seconds", d.Seconds())
}

func fallbackWalletReason(reason, fallback string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return fallback
	}
	return reason
}
