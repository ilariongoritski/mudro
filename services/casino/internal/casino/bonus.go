package casino

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrBonusAlreadyClaimed            = errors.New("bonus already claimed")
	ErrBonusVerificationRequired      = errors.New("bonus verification required")
	ErrBonusVerificationNotConfigured = errors.New("bonus verification not configured")
	ErrBonusVerificationDenied        = errors.New("bonus verification denied")
	ErrBonusVerificationUnavailable   = errors.New("bonus verification unavailable")
)

func (s *Store) GetBonusState(ctx context.Context, actor ParticipantInput, limit int) (BonusState, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return BonusState{}, err
	}
	if err := s.ensurePlayer(ctx, nil, actor, cfg); err != nil {
		return BonusState{}, err
	}

	state, err := s.loadBonusState(ctx, actor.UserID)
	if err != nil {
		return BonusState{}, err
	}

	recent, err := s.GetBonusHistory(ctx, actor.UserID, limit)
	if err != nil {
		return BonusState{}, err
	}
	state.RecentClaims = recent
	return state, nil
}

func (s *Store) GetBonusHistory(ctx context.Context, userID int64, limit int) ([]BonusClaimItem, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	rows, err := s.pool.Query(ctx, `
		select
			id,
			user_id,
			bonus_type,
			status,
			verification_status,
			verification_message,
			free_spins_granted,
			telegram_user_id,
			telegram_username,
			telegram_channel,
			metadata::text,
			claimed_at,
			created_at,
			updated_at
		from casino_bonus_claims
		where user_id = $1
		order by created_at desc, id desc
		limit $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]BonusClaimItem, 0, limit)
	for rows.Next() {
		var (
			item         BonusClaimItem
			metadataText string
		)
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.BonusKind,
			&item.Status,
			&item.VerificationStatus,
			&item.VerificationMsg,
			&item.FreeSpinsGranted,
			&item.TelegramUserID,
			&item.TelegramUsername,
			&item.TelegramChannel,
			&metadataText,
			&item.ClaimedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.Metadata = json.RawMessage(metadataText)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) ClaimSubscriptionBonus(ctx context.Context, actor ParticipantInput, initData string) (*BonusClaimResponse, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePlayer(ctx, nil, actor, cfg); err != nil {
		return nil, err
	}

	currentState, err := s.loadBonusState(ctx, actor.UserID)
	if err != nil {
		return nil, err
	}
	if currentState.Claimed {
		return &BonusClaimResponse{
			Status:             "already_claimed",
			VerificationStatus: currentState.VerificationStatus,
			State:              currentState,
		}, ErrBonusAlreadyClaimed
	}

	verification, err := verifyBonusSubscription(ctx, initData)
	if err != nil {
		return &BonusClaimResponse{
			Status:             verification.Status,
			VerificationStatus: verification.Status,
			State:              currentState,
		}, err
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

	var claimID int64
	freeSpinsGranted := BonusFreeSpins()
	if freeSpinsGranted <= 0 {
		freeSpinsGranted = DefaultBonusFreeSpins
	}
	metadata := map[string]any{
		"source":              "telegram_channel",
		"verification_status": verification.Status,
		"telegram_user_id":    verification.TelegramUserID,
		"telegram_username":   verification.TelegramUsername,
		"telegram_channel":    verification.TelegramChannel,
	}
	err = tx.QueryRow(ctx, `
		insert into casino_bonus_claims (
			user_id,
			bonus_type,
			status,
			verification_status,
			verification_message,
			free_spins_granted,
			telegram_user_id,
			telegram_username,
			telegram_channel,
			metadata,
			claimed_at,
			created_at,
			updated_at
		)
		values ($1, $2, 'claimed', $3, $4, $5, $6, $7, $8, $9::jsonb, now(), now(), now())
		on conflict (user_id, bonus_type) do nothing
		returning id
	`, actor.UserID, BonusKindSubscription, verification.Status, verification.Message, freeSpinsGranted, nullableInt64Ptr(verification.TelegramUserID), nullableStringPtr(verification.TelegramUsername), nullableStringPtr(verification.TelegramChannel), marshalJSON(metadata)).Scan(&claimID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &BonusClaimResponse{
				Status:             "already_claimed",
				VerificationStatus: currentState.VerificationStatus,
				State:              currentState,
			}, ErrBonusAlreadyClaimed
		}
		return nil, err
	}

	var updatedFreeSpins int64
	now := nowUTC()
	if err := tx.QueryRow(ctx, `
		update casino_players
		set free_spins_balance = free_spins_balance + $2,
			bonus_claim_status = 'claimed',
			bonus_claimed_at = coalesce(bonus_claimed_at, $3),
			bonus_verified_at = coalesce(bonus_verified_at, $3),
			updated_at = now()
		where user_id = $1
		returning free_spins_balance
	`, actor.UserID, freeSpinsGranted, now).Scan(&updatedFreeSpins); err != nil {
		return nil, err
	}

	if err := s.insertActivityTx(ctx, tx, actor.UserID, "bonus", fmt.Sprintf("subscription:%d", actor.UserID), 0, 0, 0, "CLAIMED", map[string]any{
		"bonus_type":          string(BonusKindSubscription),
		"bonus_claim_id":      claimID,
		"free_spins_granted":  freeSpinsGranted,
		"free_spins_balance":  updatedFreeSpins,
		"verification_status": verification.Status,
		"telegram_user_id":    verification.TelegramUserID,
		"telegram_username":   verification.TelegramUsername,
		"telegram_channel":    verification.TelegramChannel,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true

	state, err := s.GetBonusState(ctx, actor, 10)
	if err != nil {
		return nil, err
	}
	return &BonusClaimResponse{
		Status:             "claimed",
		VerificationStatus: verification.Status,
		State:              state,
	}, nil
}

func (s *Store) loadBonusState(ctx context.Context, userID int64) (BonusState, error) {
	var state BonusState
	state.UserID = userID
	state.BonusKind = BonusKindSubscription

	var (
		claimID             *int64
		claimStatus         *string
		verificationStatus  *string
		verificationMessage *string
		freeSpinsGranted    *int64
		telegramUserID      *int64
		telegramUsername    *string
		telegramChannel     *string
		claimedAt           *time.Time
		bonusClaimStatus    *string
		bonusClaimedAt      *time.Time
		bonusVerifiedAt     *time.Time
		freeSpinsBalance    int64
	)

	err := s.pool.QueryRow(ctx, `
		select
			p.free_spins_balance,
			p.bonus_claim_status,
			p.bonus_claimed_at,
			p.bonus_verified_at,
			c.id,
			c.status,
			c.verification_status,
			c.verification_message,
			c.free_spins_granted,
			c.telegram_user_id,
			c.telegram_username,
			c.telegram_channel,
			c.claimed_at
		from casino_players p
		left join lateral (
			select *
			from casino_bonus_claims c
			where c.user_id = p.user_id and c.bonus_type = $2
			order by c.id desc
			limit 1
		) c on true
		where p.user_id = $1
	`, userID, BonusKindSubscription).Scan(
		&freeSpinsBalance,
		&bonusClaimStatus,
		&bonusClaimedAt,
		&bonusVerifiedAt,
		&claimID,
		&claimStatus,
		&verificationStatus,
		&verificationMessage,
		&freeSpinsGranted,
		&telegramUserID,
		&telegramUsername,
		&telegramChannel,
		&claimedAt,
	)
	if err != nil {
		return BonusState{}, err
	}

	state.FreeSpinsBalance = freeSpinsBalance
	state.ClaimedAt = claimedAt
	state.VerifiedAt = bonusVerifiedAt
	state.Claimed = bonusClaimedAt != nil || freeSpinsBalance > 0 || (bonusClaimStatus != nil && *bonusClaimStatus != "") || (claimStatus != nil && *claimStatus != "")
	state.FreeSpinsGranted = derefInt64Ptr(freeSpinsGranted)
	state.TelegramUserID = telegramUserID
	state.TelegramUsername = firstNonEmptyPtr(telegramUsername, "")
	state.TelegramChannel = firstNonEmptyPtr(telegramChannel, BonusTelegramChannel())

	switch {
	case bonusClaimStatus != nil && *bonusClaimStatus != "":
		state.ClaimStatus = *bonusClaimStatus
	case claimStatus != nil && *claimStatus != "":
		state.ClaimStatus = *claimStatus
	case state.Claimed:
		state.ClaimStatus = "claimed"
	default:
		state.ClaimStatus = "available"
	}

	switch {
	case verificationStatus != nil && *verificationStatus != "":
		state.VerificationStatus = *verificationStatus
	case state.Claimed:
		state.VerificationStatus = "verified"
	case BonusTelegramBotToken() == "" || BonusTelegramChannel() == "":
		state.VerificationStatus = "not_configured"
	default:
		state.VerificationStatus = "verification_required"
	}
	if verificationMessage != nil {
		state.VerificationMsg = *verificationMessage
	}
	if claimID == nil {
		state.FreeSpinsGranted = 0
	}
	return state, nil
}

func firstNonEmptyPtr[T ~string](value *T, fallback T) T {
	if value != nil && *value != "" {
		return *value
	}
	return fallback
}

func derefInt64Ptr(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func nullableInt64Ptr(v int64) *int64 {
	return &v
}

func nullableStringPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
