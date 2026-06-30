package casino

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (s *Store) GetBalance(ctx context.Context, actor ParticipantInput) (int64, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return 0, err
	}
	if err := s.ensurePlayer(ctx, nil, actor, cfg); err != nil {
		return 0, err
	}
	balance, _, _, err := s.getWalletState(ctx, s.pool, actor.UserID)
	return balance, err
}

func (s *Store) GetBalanceDetails(ctx context.Context, actor ParticipantInput) (int64, int64, bool, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return 0, 0, false, err
	}
	if err := s.ensurePlayer(ctx, nil, actor, cfg); err != nil {
		return 0, 0, false, err
	}
	return s.getWalletState(ctx, s.pool, actor.UserID)
}

func (s *Store) setBalanceTx(ctx context.Context, tx pgx.Tx, userID int64, balance int64) error {
	return s.setWalletStateTx(ctx, tx, userID, balance, nil)
}

func (s *Store) setWalletStateTx(ctx context.Context, tx pgx.Tx, userID int64, balance int64, freeSpins *int64) error {
	_, err := tx.Exec(ctx, `
		update casino_players
		set balance = $2,
			free_spins_balance = coalesce($3, free_spins_balance),
			wallet_projection_source = 'microservice_projection',
			wallet_projection_note = 'projection_write',
			wallet_projection_updated_at = now(),
			updated_at = now()
		where user_id = $1
	`, userID, balance, nullableInt64(freeSpins))
	return err
}

func (s *Store) getWalletState(ctx context.Context, q queryable, userID int64) (int64, int64, bool, error) {
	var balance, freeSpins int64
	var bonusClaimed bool
	err := q.QueryRow(ctx, `
		select balance, free_spins_balance, (bonus_claimed_at is not null or bonus_claim_status <> '')
		from casino_players
		where user_id = $1
	`, userID).Scan(&balance, &freeSpins, &bonusClaimed)
	return balance, freeSpins, bonusClaimed, err
}

func (s *Store) getWalletStateForUpdate(ctx context.Context, tx pgx.Tx, userID int64) (int64, int64, string, string, int, bool, error) {
	var balance, freeSpins int64
	var bonusClaimed bool
	var clientSeed, serverSeed string
	var nonce int
	err := tx.QueryRow(ctx, `
		select balance, free_spins_balance, client_seed, coalesce(server_seed, ''), current_nonce, (bonus_claimed_at is not null or bonus_claim_status <> '')
		from casino_players
		where user_id = $1
		for update
	`, userID).Scan(&balance, &freeSpins, &clientSeed, &serverSeed, &nonce, &bonusClaimed)
	return balance, freeSpins, clientSeed, serverSeed, nonce, bonusClaimed, err
}

func (s *Store) recordTransferTx(ctx context.Context, tx pgx.Tx, kind string, userID int64, amount int64, meta map[string]any) error {
	if amount == 0 {
		return nil
	}
	var transferID string
	metadata := marshalJSON(meta)
	if err := tx.QueryRow(ctx, `
		insert into casino_transfers (kind, metadata, created_at)
		values ($1, $2, now())
		returning id
	`, kind, metadata).Scan(&transferID); err != nil {
		return err
	}

	var userAccountID string
	if err := tx.QueryRow(ctx, `select id from casino_accounts where code = $1 and type = 'user'`, fmt.Sprintf("USER_%d", userID)).Scan(&userAccountID); err != nil {
		return err
	}
	var houseAccountID string
	if err := tx.QueryRow(ctx, `select id from casino_accounts where code = $1 and type = 'system'`, "SYSTEM_HOUSE_POOL").Scan(&houseAccountID); err != nil {
		return err
	}

	directionUser := "debit"
	directionHouse := "credit"
	if kind == "bet_payout" || kind == "win" {
		directionUser = "credit"
		directionHouse = "debit"
	}

	if _, err := tx.Exec(ctx, `
		insert into casino_ledger_entries (transfer_id, account_id, direction, amount, metadata, created_at)
		values ($1, $2, $3, $4, $5, now())
	`, transferID, userAccountID, directionUser, amount, metadata); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		insert into casino_ledger_entries (transfer_id, account_id, direction, amount, metadata, created_at)
		values ($1, $2, $3, $4, $5, now())
	`, transferID, houseAccountID, directionHouse, amount, metadata); err != nil {
		return err
	}
	return nil
}
