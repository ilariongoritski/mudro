package casino

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Account struct {
	ID       string
	UserID   string
	Code     string
	Currency string
	Balance  float64
}

type Round struct {
	ID             string
	UserID         string
	ServerSeed     string
	ServerSeedHash string
	ClientSeed     string
	Nonce          int
	RoundHash      string
	Roll           *int
	BetAmount      *float64
	PayoutAmount   *float64
	Multiplier     *float64
	TierLabel      *string
	Status         string
	CreatedAt      time.Time
}

type IdempotencyKey struct {
	ID           string
	Status       string
	ResponseJSON []byte
}

// EnsureUserAccount creates the casino account for a user if it doesn't exist.
func EnsureUserAccount(ctx context.Context, pool *pgxpool.Pool, userID string, startBalance float64) (*Account, error) {
	code := "USER_" + userID

	var acct Account
	err := pool.QueryRow(ctx, `
		INSERT INTO casino_accounts (user_id, type, code, currency, balance)
		VALUES (NULL, 'user', $1, 'МДР', $2)
		ON CONFLICT (code) DO UPDATE SET updated_at = now()
		RETURNING id, COALESCE(user_id::text, ''), code, currency, balance
	`, code, startBalance).Scan(&acct.ID, &acct.UserID, &acct.Code, &acct.Currency, &acct.Balance)
	if err != nil {
		return nil, fmt.Errorf("ensure account: %w", err)
	}
	return &acct, nil
}

func GetUserAccount(ctx context.Context, pool *pgxpool.Pool, userID string) (*Account, error) {
	code := "USER_" + userID
	var acct Account
	err := pool.QueryRow(ctx, `
		SELECT id, COALESCE(user_id::text, ''), code, currency, balance
		FROM casino_accounts WHERE code = $1
	`, code).Scan(&acct.ID, &acct.UserID, &acct.Code, &acct.Currency, &acct.Balance)
	if err != nil {
		return nil, fmt.Errorf("get account %s: %w", code, err)
	}
	return &acct, nil
}

func GetSystemAccount(ctx context.Context, q pgx.Tx, code string) (*Account, error) {
	var acct Account
	err := q.QueryRow(ctx, `
		SELECT id, COALESCE(user_id::text, ''), code, currency, balance
		FROM casino_accounts WHERE code = $1
	`, code).Scan(&acct.ID, &acct.UserID, &acct.Code, &acct.Currency, &acct.Balance)
	if err != nil {
		return nil, fmt.Errorf("get system account %s: %w", code, err)
	}
	return &acct, nil
}

func PrepareRound(ctx context.Context, pool *pgxpool.Pool, userID, serverSeed, seedHash string) (*Round, error) {
	var r Round
	err := pool.QueryRow(ctx, `
		INSERT INTO casino_rounds (user_id, server_seed, server_seed_hash)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, server_seed, server_seed_hash, status, created_at
	`, userID, serverSeed, seedHash).Scan(&r.ID, &r.UserID, &r.ServerSeed, &r.ServerSeedHash, &r.Status, &r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("prepare round: %w", err)
	}
	r.Nonce = 0
	return &r, nil
}

func GetPreparedRound(ctx context.Context, tx pgx.Tx, roundID, userID string) (*Round, error) {
	var r Round
	err := tx.QueryRow(ctx, `
		SELECT id, user_id, server_seed, server_seed_hash, nonce, status
		FROM casino_rounds
		WHERE id = $1 AND user_id = $2 AND status = 'prepared'
		FOR UPDATE
	`, roundID, userID).Scan(&r.ID, &r.UserID, &r.ServerSeed, &r.ServerSeedHash, &r.Nonce, &r.Status)
	if err != nil {
		return nil, fmt.Errorf("get round: %w", err)
	}
	return &r, nil
}

func ResolveRound(ctx context.Context, tx pgx.Tx, roundID, clientSeed, roundHash string,
	nonce, roll int, betAmount, payoutAmount, multiplier float64, tierLabel string) error {
	_, err := tx.Exec(ctx, `
		UPDATE casino_rounds
		SET client_seed = $2, nonce = $3, round_hash = $4, roll = $5,
		    bet_amount = $6, payout_amount = $7, multiplier = $8, tier_label = $9,
		    status = 'resolved', resolved_at = now()
		WHERE id = $1
	`, roundID, clientSeed, nonce, roundHash, roll, betAmount, payoutAmount, multiplier, tierLabel)
	return err
}

func CreateTransfer(ctx context.Context, tx pgx.Tx, kind string,
	debitAcctID, creditAcctID string, amount float64, metadata map[string]any) error {

	var transferID string
	err := tx.QueryRow(ctx, `
		INSERT INTO casino_transfers (kind, metadata)
		VALUES ($1, $2)
		RETURNING id
	`, kind, metadata).Scan(&transferID)
	if err != nil {
		return fmt.Errorf("create transfer: %w", err)
	}

	// Debit entry
	_, err = tx.Exec(ctx, `
		INSERT INTO casino_ledger_entries (transfer_id, account_id, direction, amount)
		VALUES ($1, $2, 'debit', $3)
	`, transferID, debitAcctID, amount)
	if err != nil {
		return fmt.Errorf("debit entry: %w", err)
	}

	// Credit entry
	_, err = tx.Exec(ctx, `
		INSERT INTO casino_ledger_entries (transfer_id, account_id, direction, amount)
		VALUES ($1, $2, 'credit', $3)
	`, transferID, creditAcctID, amount)
	if err != nil {
		return fmt.Errorf("credit entry: %w", err)
	}

	// Update balances
	_, err = tx.Exec(ctx, `UPDATE casino_accounts SET balance = balance - $2, updated_at = now() WHERE id = $1`, debitAcctID, amount)
	if err != nil {
		return fmt.Errorf("update debit balance: %w", err)
	}

	_, err = tx.Exec(ctx, `UPDATE casino_accounts SET balance = balance + $2, updated_at = now() WHERE id = $1`, creditAcctID, amount)
	if err != nil {
		return fmt.Errorf("update credit balance: %w", err)
	}

	return nil
}

func AcquireIdempotencyKey(ctx context.Context, tx pgx.Tx, userID, key, hash string) (*IdempotencyKey, error) {
	var ik IdempotencyKey
	err := tx.QueryRow(ctx, `
		INSERT INTO casino_idempotency (user_id, key, request_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, key) DO UPDATE SET user_id = casino_idempotency.user_id
		RETURNING id, status, response
	`, userID, key, hash).Scan(&ik.ID, &ik.Status, &ik.ResponseJSON)
	if err != nil {
		return nil, fmt.Errorf("idempotency: %w", err)
	}
	return &ik, nil
}

func CompleteIdempotencyKey(ctx context.Context, tx pgx.Tx, id string, response []byte) error {
	_, err := tx.Exec(ctx, `
		UPDATE casino_idempotency SET status = 'succeeded', response = $2 WHERE id = $1
	`, id, response)
	return err
}
