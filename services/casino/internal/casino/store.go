package casino

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrUnauthorized        = errors.New("unauthorized")
)

type Store struct {
	pool   *pgxpool.Pool
	engine *Engine
}

func NewStore(pool *pgxpool.Pool, engine *Engine) *Store {
	return &Store{pool: pool, engine: engine}
}

func (s *Store) EnsureSeedConfig(ctx context.Context) error {
	_, err := s.GetConfig(ctx)
	return err
}

func (s *Store) Health(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Store) GetConfig(ctx context.Context) (Config, error) {
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
			return s.GetConfig(ctx)
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
	return err
}

func (s *Store) GetBalance(ctx context.Context, actor ParticipantInput) (int64, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return 0, err
	}
	if err := s.ensureParticipant(ctx, nil, actor, cfg); err != nil {
		return 0, err
	}
	var balance int64
	err = s.pool.QueryRow(ctx, `select coins from casino_participants where user_id = $1`, actor.UserID).Scan(&balance)
	return balance, err
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
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	cfg, err := s.getConfigTx(ctx, tx)
	if err != nil {
		return nil, err
	}
	if err := s.ensureParticipant(ctx, tx, actor, cfg); err != nil {
		return nil, err
	}

	var balance int64
	if err := tx.QueryRow(ctx, `select coins from casino_participants where user_id = $1 for update`, actor.UserID).Scan(&balance); err != nil {
		return nil, err
	}
	if balance < bet {
		return nil, ErrInsufficientBalance
	}

	symbols, win, err := s.engine.Spin(cfg, bet)
	if err != nil {
		return nil, err
	}
	newBalance := balance - bet + win

	if _, err := tx.Exec(ctx, `
		update casino_participants
		set coins = $2,
			spins_count = spins_count + 1,
			last_spin_at = now(),
			updated_at = now()
		where user_id = $1
	`, actor.UserID, newBalance); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		insert into casino_spins (user_id, bet, win, symbols)
		values ($1, $2, $3, $4::jsonb)
	`, actor.UserID, bet, win, marshalJSON(symbols)); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &SpinResult{
		Balance: newBalance,
		Symbols: symbols,
		Win:     win,
	}, nil
}

func (s *Store) ensureParticipant(ctx context.Context, tx pgx.Tx, actor ParticipantInput, cfg Config) error {
	if actor.UserID <= 0 {
		return ErrUnauthorized
	}

	query := `
		insert into casino_participants (
			user_id, username, email, role, coins, updated_at
		)
		values ($1, $2, $3, $4, $5, now())
		on conflict (user_id) do update set
			username = excluded.username,
			email = excluded.email,
			role = excluded.role,
			updated_at = now()
	`
	args := []any{
		actor.UserID,
		actor.Username,
		actor.Email,
		actor.Role,
		cfg.InitialBalance,
	}

	if tx != nil {
		_, err := tx.Exec(ctx, query, args...)
		return err
	}

	_, err := s.pool.Exec(ctx, query, args...)
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
