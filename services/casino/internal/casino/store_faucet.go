package casino

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrFaucetCooldown = errors.New("faucet on cooldown")
	ErrFaucetNotConfigured = errors.New("faucet amount not configured")
)

type FaucetResponse struct {
	Claimed       bool       `json:"claimed"`
	Amount        int64      `json:"amount"`
	Balance       int64      `json:"balance"`
	NextClaimAt   *time.Time `json:"next_claim_at,omitempty"`
	CooldownHours int        `json:"cooldown_hours"`
}

func (s *Store) ClaimFaucet(ctx context.Context, actor ParticipantInput) (*FaucetResponse, error) {
	amount := FaucetAmount()
	if amount <= 0 {
		return nil, ErrFaucetNotConfigured
	}
	cooldown := FaucetCooldown()

	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return nil, err
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

	if err := s.ensurePlayer(ctx, tx, actor, cfg); err != nil {
		return nil, err
	}

	var lastClaim *time.Time
	err = tx.QueryRow(ctx, `select last_faucet_claim from casino_players where user_id = $1 for update`, actor.UserID).Scan(&lastClaim)
	if err != nil {
		return nil, err
	}

	now := nowUTC()
	var nextClaimAt *time.Time
	if lastClaim != nil {
		elapsed := now.Sub(*lastClaim)
		if elapsed < cooldown {
			nc := lastClaim.Add(cooldown)
			nextClaimAt = &nc
			return &FaucetResponse{
				Claimed:       false,
				Amount:        amount,
				NextClaimAt:   nextClaimAt,
				CooldownHours: int(cooldown.Hours()),
			}, ErrFaucetCooldown
		}
	}

	var balance int64
	err = tx.QueryRow(ctx, `select balance from casino_players where user_id = $1 for update`, actor.UserID).Scan(&balance)
	if err != nil {
		return nil, err
	}

	newBalance := balance + amount
	if _, err := tx.Exec(ctx, `update casino_players set balance = $2, last_faucet_claim = now(), updated_at = now() where user_id = $1`, actor.UserID, newBalance); err != nil {
		return nil, err
	}

	if err := s.recordTransferTx(ctx, tx, "faucet_claim", actor.UserID, amount, map[string]any{"game": "faucet"}); err != nil {
		return nil, err
	}

	faucetRef := fmt.Sprintf("faucet_%d_%d", actor.UserID, now.Unix())
	if err := s.insertActivityTx(ctx, tx, actor.UserID, "faucet", faucetRef, 0, amount, amount, "CLAIMED", map[string]any{
		"amount":  amount,
		"source":  "daily_faucet",
	}); err != nil {
		return nil, err
	}

	if err := s.enqueueBalanceSyncTx(ctx, tx, actor.UserID, "faucet_claim", &newBalance); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true

	return &FaucetResponse{
		Claimed:       true,
		Amount:        amount,
		Balance:       newBalance,
		CooldownHours: int(cooldown.Hours()),
	}, nil
}

func (s *Store) GetFaucetState(ctx context.Context, actor ParticipantInput) (*FaucetResponse, error) {
	amount := FaucetAmount()
	cooldown := FaucetCooldown()

	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePlayer(ctx, nil, actor, cfg); err != nil {
		return nil, err
	}

	var lastClaim *time.Time
	err = s.pool.QueryRow(ctx, `select last_faucet_claim from casino_players where user_id = $1`, actor.UserID).Scan(&lastClaim)
	if err != nil {
		return nil, err
	}

	now := nowUTC()
	var nextClaimAt *time.Time
	claimed := false
	if lastClaim != nil {
		elapsed := now.Sub(*lastClaim)
		if elapsed < cooldown {
			nc := lastClaim.Add(cooldown)
			nextClaimAt = &nc
		} else {
			claimed = true
		}
	} else {
		claimed = true
	}

	return &FaucetResponse{
		Claimed:       claimed,
		Amount:        amount,
		NextClaimAt:   nextClaimAt,
		CooldownHours: int(cooldown.Hours()),
	}, nil
}
