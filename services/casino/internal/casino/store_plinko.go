package casino

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

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
