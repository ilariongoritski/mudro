package casino

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)



type pgCasinoRepository struct {
	pool *pgxpool.Pool
}

func NewPgRepository(pool *pgxpool.Pool) CasinoRepository {
	return &pgCasinoRepository{pool: pool}
}

// EnsureUserAccount creates the casino account for a user if it doesn't exist.
func (repo *pgCasinoRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return repo.pool.Begin(ctx)
}

func (repo *pgCasinoRepository) EnsureUserAccount(ctx context.Context, userID string, startBalance float64) (*Account, error) {
	code := "USER_" + userID

	var acct Account
	err := repo.pool.QueryRow(ctx, `
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

func (repo *pgCasinoRepository) GetUserAccount(ctx context.Context, userID string) (*Account, error) {
	code := "USER_" + userID
	var acct Account
	err := repo.pool.QueryRow(ctx, `
		SELECT id, COALESCE(user_id::text, ''), code, currency, balance
		FROM casino_accounts WHERE code = $1
	`, code).Scan(&acct.ID, &acct.UserID, &acct.Code, &acct.Currency, &acct.Balance)
	if err != nil {
		return nil, fmt.Errorf("get account %s: %w", code, err)
	}
	return &acct, nil
}

func (repo *pgCasinoRepository) GetSystemAccount(ctx context.Context, q pgx.Tx, code string) (*Account, error) {
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

func (repo *pgCasinoRepository) PrepareRound(ctx context.Context, userID, serverSeed, seedHash string) (*Round, error) {
	var rnd Round
	err := repo.pool.QueryRow(ctx, `
		INSERT INTO casino_rounds (user_id, server_seed, server_seed_hash)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, server_seed, server_seed_hash, status, created_at
	`, userID, serverSeed, seedHash).Scan(&rnd.ID, &rnd.UserID, &rnd.ServerSeed, &rnd.ServerSeedHash, &rnd.Status, &rnd.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("prepare round: %w", err)
	}
	rnd.Nonce = 0
	return &rnd, nil
}

func (repo *pgCasinoRepository) GetPreparedRound(ctx context.Context, tx pgx.Tx, roundID, userID string) (*Round, error) {
	var rnd Round
	err := tx.QueryRow(ctx, `
		SELECT id, user_id, server_seed, server_seed_hash, nonce, status
		FROM casino_rounds
		WHERE id = $1 AND user_id = $2 AND status = 'prepared'
		FOR UPDATE
	`, roundID, userID).Scan(&rnd.ID, &rnd.UserID, &rnd.ServerSeed, &rnd.ServerSeedHash, &rnd.Nonce, &rnd.Status)
	if err != nil {
		return nil, fmt.Errorf("get round: %w", err)
	}
	return &rnd, nil
}

func (repo *pgCasinoRepository) ResolveRound(ctx context.Context, tx pgx.Tx, roundID, clientSeed, roundHash string,
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

func (repo *pgCasinoRepository) CreateTransfer(ctx context.Context, tx pgx.Tx, kind string,
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

func (repo *pgCasinoRepository) AcquireIdempotencyKey(ctx context.Context, tx pgx.Tx, userID, key, hash string) (*IdempotencyKey, error) {
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

func (repo *pgCasinoRepository) CompleteIdempotencyKey(ctx context.Context, tx pgx.Tx, id string, response []byte) error {
	_, err := tx.Exec(ctx, `
		UPDATE casino_idempotency SET status = 'succeeded', response = $2 WHERE id = $1
	`, id, response)
	return err
}

func (repo *pgCasinoRepository) GetStats(ctx context.Context) (userCount, roundCount int, houseBalance, totalBet, totalPayout float64, err error) {
	_ = repo.pool.QueryRow(ctx, `SELECT count(*) FROM casino_accounts WHERE type = 'user'`).Scan(&userCount)
	_ = repo.pool.QueryRow(ctx, `SELECT count(*) FROM casino_rounds WHERE status = 'resolved'`).Scan(&roundCount)
	_ = repo.pool.QueryRow(ctx, `SELECT COALESCE(balance, 0) FROM casino_accounts WHERE code = $1`, HouseAccountCode).Scan(&houseBalance)
	_ = repo.pool.QueryRow(ctx, `SELECT COALESCE(SUM(bet_amount), 0), COALESCE(SUM(payout_amount), 0) FROM casino_rounds WHERE status = 'resolved'`).Scan(&totalBet, &totalPayout)
	return
}

func (repo *pgCasinoRepository) GetRTPProfiles(ctx context.Context) ([]map[string]any, error) {
	rows, err := repo.pool.Query(ctx, `SELECT id, name, rtp, paytable, is_default FROM casino_rtp_profiles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []map[string]any
	for rows.Next() {
		var id, name string
		var rtp float64
		var paytable []byte
		var isDefault bool
		if err := rows.Scan(&id, &name, &rtp, &paytable, &isDefault); err == nil {
			profiles = append(profiles, map[string]any{
				"id": id, "name": name, "rtp": rtp, "paytable": json.RawMessage(paytable), "isDefault": isDefault,
			})
		}
	}
	return profiles, nil
}

func (repo *pgCasinoRepository) UpsertRTPProfile(ctx context.Context, name string, rtp float64, paytable []byte, isDefault bool) (string, error) {
	if isDefault {
		_, _ = repo.pool.Exec(ctx, `UPDATE casino_rtp_profiles SET is_default = false WHERE is_default = true`)
	}
	var id string
	err := repo.pool.QueryRow(ctx, `
		INSERT INTO casino_rtp_profiles (name, rtp, paytable, is_default)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name) DO UPDATE SET rtp = $2, paytable = $3, is_default = $4, updated_at = now()
		RETURNING id
	`, name, rtp, paytable, isDefault).Scan(&id)
	return id, err
}

func (repo *pgCasinoRepository) DeleteRTPProfile(ctx context.Context, id string) error {
	_, err := repo.pool.Exec(ctx, `DELETE FROM casino_rtp_profiles WHERE id = $1`, id)
	return err
}

func (repo *pgCasinoRepository) GetUsers(ctx context.Context, limit int) ([]map[string]any, error) {
	rows, err := repo.pool.Query(ctx, `SELECT id, code, currency, balance FROM casino_accounts WHERE type = 'user' ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []map[string]any
	for rows.Next() {
		var id, code, currency string
		var balance float64
		if err := rows.Scan(&id, &code, &currency, &balance); err == nil {
			users = append(users, map[string]any{"id": id, "code": code, "currency": currency, "balance": balance})
		}
	}
	return users, nil
}

func (repo *pgCasinoRepository) GetActiveRtpProfile(ctx context.Context, userID string) (*RtpProfile, error) {
	var id, name string
	var rtp float64
	var paytableJSON []byte
	var isDefault bool

	row := repo.pool.QueryRow(ctx, `
		SELECT p.id, p.name, p.rtp, p.paytable, p.is_default
		FROM casino_rtp_assignments a
		JOIN casino_rtp_profiles p ON p.id = a.rtp_profile_id
		WHERE a.user_id = $1 AND (a.expires_at IS NULL OR a.expires_at > now())
		ORDER BY a.created_at DESC
		LIMIT 1
	`, userID)
	if err := row.Scan(&id, &name, &rtp, &paytableJSON, &isDefault); err != nil {
		row = repo.pool.QueryRow(ctx, `
			SELECT id, name, rtp, paytable, is_default
			FROM casino_rtp_profiles
			WHERE is_default = true
			LIMIT 1
		`)
		if err := row.Scan(&id, &name, &rtp, &paytableJSON, &isDefault); err != nil {
			return nil, fmt.Errorf("no default RTP profile: %w", err)
		}
	}

	var tiers []PaytableTier
	if err := json.Unmarshal(paytableJSON, &tiers); err != nil {
		return nil, fmt.Errorf("parse paytable: %w", err)
	}

	return &RtpProfile{
		ID:        id,
		Name:      name,
		Rtp:       rtp,
		Paytable:  tiers,
		IsDefault: isDefault,
	}, nil
}

