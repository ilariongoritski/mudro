package casino

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// Service forms the usecase layer for casino operations.
type Service struct {
	repo CasinoRepository
}

// NewService creates a new casino usecase service.
func NewService(repo CasinoRepository) *Service {
	return &Service{repo: repo}
}

// PrepareRound initiates a new provably fair game round.
func (s *Service) PrepareRound(ctx context.Context, userID string) (*Round, error) {
	serverSeed := GenerateServerSeed()
	seedHash := HashServerSeed(serverSeed)

	return s.repo.PrepareRound(ctx, userID, serverSeed, seedHash)
}

// EnsureUserAccount ensures a wallet is created for the new user.
func (s *Service) EnsureUserAccount(ctx context.Context, userID string, startBalance float64) (*Account, error) {
	return s.repo.EnsureUserAccount(ctx, userID, startBalance)
}

// GetBalance gets the current player's balance.
func (s *Service) GetBalance(ctx context.Context, userID string) (*Account, error) {
	return s.repo.GetUserAccount(ctx, userID)
}

// GrantFaucet gives the user free coins (testnet/demo mode).
func (s *Service) GrantFaucet(ctx context.Context, userID string, amount float64) (*FaucetResult, error) {
	if amount <= 0 {
		return nil, errors.New("faucet amount must be positive")
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	userAcct, err := s.repo.GetSystemAccount(ctx, tx, "USER_"+userID)
	if err != nil {
		return nil, err
	}
	houseAcct, err := s.repo.GetSystemAccount(ctx, tx, HouseAccountCode)
	if err != nil {
		return nil, err
	}

	err = s.repo.CreateTransfer(ctx, tx, "adjustment", houseAcct.ID, userAcct.ID, amount, map[string]any{
		"reason": "DEMO_FAUCET",
	})
	if err != nil {
		return nil, fmt.Errorf("faucet transfer: %w", err)
	}

	// Read strictly to get fresh balance inside tx
	freshAcct, err := s.repo.GetSystemAccount(ctx, tx, "USER_"+userID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &FaucetResult{Amount: amount, Balance: freshAcct.Balance}, nil
}

// PlaceBet resolves a game round with the user's client seed and calculates payout.
func (s *Service) PlaceBet(ctx context.Context, input BetInput) (*BetResult, error) {
	if input.BetAmount <= 0 {
		return nil, errors.New("bet amount must be positive")
	}

	rtpProfile, err := GetActiveRtpProfile(ctx, s.repo, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("rtp profile: %w", err)
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Idempotency check
	idem, err := s.repo.AcquireIdempotencyKey(ctx, tx, input.UserID, input.IdempotencyKey,
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

	round, err := s.repo.GetPreparedRound(ctx, tx, input.RoundID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("get round: %w", err)
	}

	userAcct, err := s.repo.GetSystemAccount(ctx, tx, "USER_"+input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user account: %w", err)
	}
	houseAcct, err := s.repo.GetSystemAccount(ctx, tx, HouseAccountCode)
	if err != nil {
		return nil, fmt.Errorf("house account: %w", err)
	}

	if userAcct.Balance < input.BetAmount {
		return nil, errors.New("insufficient balance")
	}

	nonce := round.Nonce + 1
	roll, roundHash := Resolve(round.ServerSeed, input.ClientSeed, nonce)
	payout := EvaluatePayout(roll, input.BetAmount, rtpProfile.Paytable)

	err = s.repo.CreateTransfer(ctx, tx, "bet_stake", userAcct.ID, houseAcct.ID, input.BetAmount, map[string]any{
		"roundId": input.RoundID, "roll": roll,
	})
	if err != nil {
		return nil, fmt.Errorf("stake transfer: %w", err)
	}

	if payout.Amount > 0 {
		err = s.repo.CreateTransfer(ctx, tx, "bet_payout", houseAcct.ID, userAcct.ID, payout.Amount, map[string]any{
			"roundId": input.RoundID, "multiplier": payout.Multiplier, "tier": payout.Label,
		})
		if err != nil {
			return nil, fmt.Errorf("payout transfer: %w", err)
		}
	}

	err = s.repo.ResolveRound(ctx, tx, input.RoundID, input.ClientSeed, roundHash,
		nonce, roll, input.BetAmount, payout.Amount, payout.Multiplier, payout.Label)
	if err != nil {
		return nil, fmt.Errorf("resolve round: %w", err)
	}

	freshAcct, err := s.repo.GetSystemAccount(ctx, tx, "USER_"+input.UserID)
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
		Balance:        freshAcct.Balance,
		Nonce:          nonce,
		RoundHash:      roundHash,
		ServerSeedHash: round.ServerSeedHash,
		ServerSeed:     round.ServerSeed,
	}

	resJSON, _ := json.Marshal(result)
	_ = s.repo.CompleteIdempotencyKey(ctx, tx, idem.ID, resJSON)

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}


func (s *Service) GetStats(ctx context.Context) (int, int, float64, float64, float64, error) {
	return s.repo.GetStats(ctx)
}
func (s *Service) GetRTPProfiles(ctx context.Context) ([]map[string]any, error) {
	return s.repo.GetRTPProfiles(ctx)
}
func (s *Service) UpsertRTPProfile(ctx context.Context, name string, rtp float64, paytable []byte, isDefault bool) (string, error) {
	return s.repo.UpsertRTPProfile(ctx, name, rtp, paytable, isDefault)
}
func (s *Service) DeleteRTPProfile(ctx context.Context, id string) error {
	return s.repo.DeleteRTPProfile(ctx, id)
}
func (s *Service) GetUsers(ctx context.Context, limit int) ([]map[string]any, error) {
	return s.repo.GetUsers(ctx, limit)
}
