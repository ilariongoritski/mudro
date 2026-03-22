package casino

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	HouseAccountCode      = "SYSTEM_HOUSE_POOL"
	SettlementAccountCode = "SYSTEM_SETTLEMENT_POOL"
)

type BetInput struct {
	UserID         string
	RoundID        string
	BetAmount      float64
	ClientSeed     string
	IdempotencyKey string
}

type BetResult struct {
	RoundID        string  `json:"roundId"`
	Roll           int     `json:"roll"`
	BetAmount      float64 `json:"betAmount"`
	PayoutAmount   float64 `json:"payoutAmount"`
	Multiplier     float64 `json:"multiplier"`
	TierLabel      string  `json:"tierLabel"`
	TierSymbol     string  `json:"tierSymbol"`
	Balance        float64 `json:"balance"`
	Nonce          int     `json:"nonce"`
	RoundHash      string  `json:"roundHash"`
	ServerSeedHash string  `json:"serverSeedHash"`
	ServerSeed     string  `json:"serverSeed"`
	FromCache      bool    `json:"fromCache"`
}

type FaucetResult struct {
	Amount  float64 `json:"amount"`
	Balance float64 `json:"balance"`
}

func PlaceBet(ctx context.Context, pool *pgxpool.Pool, input BetInput) (*BetResult, error) {
	if input.BetAmount <= 0 {
		return nil, errors.New("bet amount must be positive")
	}

	// Get RTP profile
	rtpProfile, err := GetActiveRtpProfile(ctx, pool, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("rtp profile: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Idempotency check
	idem, err := AcquireIdempotencyKey(ctx, tx, input.UserID, input.IdempotencyKey,
		fmt.Sprintf("%s:%s:%.8f", input.RoundID, input.ClientSeed, input.BetAmount))
	if err != nil {
		return nil, err
	}
	if idem.Status == "succeeded" && idem.ResponseJSON != nil {
		var cached BetResult
		if err := json.Unmarshal(idem.ResponseJSON, &cached); err == nil {
			cached.FromCache = true
			return &cached, nil
		}
	}

	// Get round
	round, err := GetPreparedRound(ctx, tx, input.RoundID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("get round: %w", err)
	}

	// Get accounts
	userAcct, err := GetSystemAccount(ctx, tx, "USER_"+input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user account: %w", err)
	}
	houseAcct, err := GetSystemAccount(ctx, tx, HouseAccountCode)
	if err != nil {
		return nil, fmt.Errorf("house account: %w", err)
	}

	if userAcct.Balance < input.BetAmount {
		return nil, errors.New("insufficient balance")
	}

	// Resolve provably fair
	nonce := round.Nonce + 1
	roll, roundHash := Resolve(round.ServerSeed, input.ClientSeed, nonce)
	payout := EvaluatePayout(roll, input.BetAmount, rtpProfile.Paytable)

	// Stake transfer: user → house
	err = CreateTransfer(ctx, tx, "bet_stake", userAcct.ID, houseAcct.ID, input.BetAmount, map[string]any{
		"roundId": input.RoundID, "roll": roll,
	})
	if err != nil {
		return nil, fmt.Errorf("stake transfer: %w", err)
	}

	// Payout transfer: house → user (if won)
	if payout.Amount > 0 {
		err = CreateTransfer(ctx, tx, "bet_payout", houseAcct.ID, userAcct.ID, payout.Amount, map[string]any{
			"roundId": input.RoundID, "multiplier": payout.Multiplier, "tier": payout.Label,
		})
		if err != nil {
			return nil, fmt.Errorf("payout transfer: %w", err)
		}
	}

	// Resolve round
	err = ResolveRound(ctx, tx, input.RoundID, input.ClientSeed, roundHash,
		nonce, roll, input.BetAmount, payout.Amount, payout.Multiplier, payout.Label)
	if err != nil {
		return nil, fmt.Errorf("resolve round: %w", err)
	}

	// Get fresh balance
	var freshBalance float64
	err = tx.QueryRow(ctx, `SELECT balance FROM casino_accounts WHERE id = $1`, userAcct.ID).Scan(&freshBalance)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	result := &BetResult{
		RoundID:        input.RoundID,
		Roll:           roll,
		BetAmount:      input.BetAmount,
		PayoutAmount:   payout.Amount,
		Multiplier:     payout.Multiplier,
		TierLabel:      payout.Label,
		TierSymbol:     payout.Symbol,
		Balance:        freshBalance,
		Nonce:          nonce,
		RoundHash:      roundHash,
		ServerSeedHash: round.ServerSeedHash,
		ServerSeed:     round.ServerSeed,
	}

	// Save idempotency
	resJSON, _ := json.Marshal(result)
	_ = CompleteIdempotencyKey(ctx, tx, idem.ID, resJSON)

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

func GrantFaucet(ctx context.Context, pool *pgxpool.Pool, userID string, amount float64) (*FaucetResult, error) {
	if amount <= 0 {
		return nil, errors.New("faucet amount must be positive")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	userAcct, err := GetSystemAccount(ctx, tx, "USER_"+userID)
	if err != nil {
		return nil, err
	}
	houseAcct, err := GetSystemAccount(ctx, tx, HouseAccountCode)
	if err != nil {
		return nil, err
	}

	err = CreateTransfer(ctx, tx, "adjustment", houseAcct.ID, userAcct.ID, amount, map[string]any{
		"reason": "DEMO_FAUCET",
	})
	if err != nil {
		return nil, fmt.Errorf("faucet transfer: %w", err)
	}

	var freshBalance float64
	err = tx.QueryRow(ctx, `SELECT balance FROM casino_accounts WHERE id = $1`, userAcct.ID).Scan(&freshBalance)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &FaucetResult{Amount: amount, Balance: freshBalance}, nil
}
