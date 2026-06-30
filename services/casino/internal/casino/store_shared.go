package casino

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
)

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

func (s *Store) insertActivityTx(ctx context.Context, tx pgx.Tx, userID int64, gameType, gameRef string, betAmount, payoutAmount, netResult int64, status string, metadata any) error {
	_, err := tx.Exec(ctx, `
		insert into casino_game_activity (
			user_id, game_type, game_ref, bet_amount, payout_amount, net_result, status, metadata
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8::jsonb)
	`, userID, gameType, gameRef, betAmount, payoutAmount, netResult, status, marshalJSON(metadata))
	return err
}

type queryable interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func nullableInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}
