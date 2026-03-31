package casino

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// CasinoRepository defines data access methods for the casino module.
type CasinoRepository interface {
	// Transactions
	BeginTx(ctx context.Context) (pgx.Tx, error)

	// Accounts
	EnsureUserAccount(ctx context.Context, userID string, startBalance float64) (*Account, error)
	GetUserAccount(ctx context.Context, userID string) (*Account, error)
	GetSystemAccount(ctx context.Context, q pgx.Tx, code string) (*Account, error)
	
	// Rounds
	PrepareRound(ctx context.Context, userID, serverSeed, seedHash string) (*Round, error)
	GetPreparedRound(ctx context.Context, tx pgx.Tx, roundID, userID string) (*Round, error)
	ResolveRound(ctx context.Context, tx pgx.Tx, roundID, clientSeed, roundHash string, nonce, roll int, betAmount, payoutAmount, multiplier float64, tierLabel string) error
	
	// Ledger
	CreateTransfer(ctx context.Context, tx pgx.Tx, kind string, debitAcctID, creditAcctID string, amount float64, metadata map[string]any) error
	
	// Idempotency
	AcquireIdempotencyKey(ctx context.Context, tx pgx.Tx, userID, key, hash string) (*IdempotencyKey, error)
	CompleteIdempotencyKey(ctx context.Context, tx pgx.Tx, id string, response []byte) error

	// Admin
	GetStats(ctx context.Context) (userCount, roundCount int, houseBalance, totalBet, totalPayout float64, err error)
	GetRTPProfiles(ctx context.Context) ([]map[string]any, error)
	GetActiveRtpProfile(ctx context.Context, userID string) (*RtpProfile, error)
	UpsertRTPProfile(ctx context.Context, name string, rtp float64, paytable []byte, isDefault bool) (string, error)
	DeleteRTPProfile(ctx context.Context, id string) error
	GetUsers(ctx context.Context, limit int) ([]map[string]any, error)
}
