package casino

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
)

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

	// Sweet Bonanza outcomes are generated entirely on the server from the
	// existing provably-fair server seed, client seed and nonce.
	sweet := newSweetEngine(serverSeed, clientSeed, int64(nonce)).Spin(bet, freeSpinUsed)
	win := sweet.TotalWin
	if sweet.FreeSpinsAwarded > 0 {
		freeSpins += sweet.FreeSpinsAwarded
	}
	// Preserve the legacy history shape while the client consumes the richer
	// sweet_bonanza timeline from the response.
	symbols := []string{"sweet_bonanza"}

	newBalance := balance + win
	if !freeSpinUsed {
		newBalance -= bet
	}

	if err := s.setWalletStateTx(ctx, tx, actor.UserID, newBalance, &freeSpins); err != nil {
		return nil, err
	}

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
		"free_spins_awarded": sweet.FreeSpinsAwarded,
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
		Config:           cfg,
		Symbols:          symbols,
		Win:              win,
		SweetBonanza:     &sweet,
	}, nil
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

// BuyBonus charges bonusBuyPrice(bet) from the player's real balance and
// credits BonusFreeSpins() free spins in one transaction. Server-authoritative:
// the client never mints free spins locally.
func (s *Store) BuyBonus(ctx context.Context, actor ParticipantInput, bet int64) (*SpinResult, error) {
	if bet <= 0 {
		return nil, fmt.Errorf("bet must be positive")
	}
	price := bonusBuyPrice(bet)

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

	balance, freeSpins, _, _, _, _, err := s.getWalletStateForUpdate(ctx, tx, actor.UserID)
	if err != nil {
		return nil, err
	}
	if freeSpins > 0 {
		return nil, fmt.Errorf("bonus already in progress")
	}
	if balance < price {
		return nil, ErrInsufficientBalance
	}

	newBalance := balance - price
	freeSpins += BonusFreeSpins()
	if err := s.setWalletStateTx(ctx, tx, actor.UserID, newBalance, &freeSpins); err != nil {
		return nil, err
	}
	if err := s.recordTransferTx(ctx, tx, "bonus_buy", actor.UserID, price, map[string]any{"bet": bet}); err != nil {
		return nil, err
	}
	if err := s.insertActivityTx(ctx, tx, actor.UserID, "slots", "bonus_buy", price, 0, -price, "BONUS BUY", map[string]any{
		"bet":                bet,
		"free_spins_balance": freeSpins,
	}); err != nil {
		return nil, err
	}
	if err := s.enqueueBalanceSyncTx(ctx, tx, actor.UserID, "bonus_buy", &newBalance); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true

	return &SpinResult{
		Balance:          newBalance,
		FreeSpinsBalance: freeSpins,
		Config:           cfg,
	}, nil
}

// bonusBuyPrice mirrors the client's BONUS_BUY_MULT (100x bet).
func bonusBuyPrice(bet int64) int64 {
	return bet * 100
}
