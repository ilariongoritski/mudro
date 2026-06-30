package casino

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (s *Store) GetProfile(ctx context.Context, actor ParticipantInput, limit int) (PlayerProfile, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return PlayerProfile{}, err
	}
	if err := s.ensurePlayer(ctx, nil, actor, cfg); err != nil {
		return PlayerProfile{}, err
	}

	var profile PlayerProfile
	err = s.pool.QueryRow(ctx, `
		select
			user_id,
			username,
			display_name,
			avatar_url,
			balance,
			free_spins_balance,
			bonus_claim_status,
			bonus_claimed_at,
			bonus_verified_at,
			total_wagered,
			total_won,
			games_played,
			roulette_rounds_played,
			level,
			xp_progress,
			last_game_at,
			client_seed,
			current_nonce,
			server_seed_hash
		from casino_players
		where user_id = $1
	`, actor.UserID).Scan(
		&profile.UserID,
		&profile.Username,
		&profile.DisplayName,
		&profile.AvatarURL,
		&profile.Balance,
		&profile.FreeSpinsBalance,
		&profile.BonusClaimStatus,
		&profile.BonusClaimedAt,
		&profile.BonusVerifiedAt,
		&profile.TotalWagered,
		&profile.TotalWon,
		&profile.GamesPlayed,
		&profile.RouletteRoundsPlayed,
		&profile.Level,
		&profile.XPProgress,
		&profile.LastGameAt,
		&profile.ClientSeed,
		&profile.CurrentNonce,
		&profile.ServerSeedHash,
	)
	if err == nil {
		profile.ProgressTarget = levelStepWagered
	}
	if err != nil {
		return PlayerProfile{}, err
	}
	profile.ProgressTarget = levelStepWagered
	if strings.TrimSpace(profile.DisplayName) == "" {
		profile.DisplayName = profile.Username
	}

	activity, err := s.GetActivity(ctx, actor.UserID, limit)
	if err != nil {
		return PlayerProfile{}, err
	}
	profile.RecentActivity = activity
	return profile, nil
}

func (s *Store) ensurePlayer(ctx context.Context, tx pgx.Tx, actor ParticipantInput, cfg Config) error {
	if actor.UserID <= 0 {
		return ErrUnauthorized
	}

	initialBalance := cfg.InitialBalance
	projectionSource := "microservice_projection"
	projectionNote := ""
	var projectionSyncedAt any
	if s.mainPool != nil {
		if mainBalance, err := s.getMainWalletBalance(ctx, actor.UserID); err == nil {
			initialBalance = mainBalance
			projectionSource = "main_wallet_bootstrap"
			projectionNote = "main_wallet_bootstrap"
			projectionSyncedAt = nowUTC()
		}
	}

	queryPlayer := `
		insert into casino_players (
			user_id,
			username,
			email,
			role,
			display_name,
			balance,
			level,
			xp_progress,
			wallet_projection_source,
			wallet_projection_note,
			wallet_projection_updated_at,
			wallet_projection_synced_at,
			updated_at,
			client_seed,
			current_nonce,
			server_seed_hash
		)
		values ($1, $2, $3, $4, $5, $6, 1, 0, $7, $8, now(), $9, now(), 'default', 0, $10)
		on conflict (user_id) do update set
			username = excluded.username,
			email = excluded.email,
			role = excluded.role,
			display_name = case when casino_players.display_name = '' then excluded.display_name else casino_players.display_name end,
			balance = case
				when casino_players.wallet_projection_synced_at is null and excluded.wallet_projection_synced_at is not null
					then excluded.balance
				else casino_players.balance
			end,
			wallet_projection_source = case
				when casino_players.wallet_projection_synced_at is null and excluded.wallet_projection_synced_at is not null
					then excluded.wallet_projection_source
				else casino_players.wallet_projection_source
			end,
			wallet_projection_note = case
				when casino_players.wallet_projection_synced_at is null and excluded.wallet_projection_synced_at is not null
					then excluded.wallet_projection_note
				else casino_players.wallet_projection_note
			end,
			wallet_projection_updated_at = now(),
			wallet_projection_synced_at = coalesce(casino_players.wallet_projection_synced_at, excluded.wallet_projection_synced_at),
			updated_at = now()
	`
	seed, _ := GenerateServerSeed()
	seedHash := HashServerSeed(seed)

	args := []any{
		actor.UserID,
		actor.Username,
		actor.Email,
		actor.Role,
		actor.Username,
		initialBalance,
		projectionSource,
		projectionNote,
		projectionSyncedAt,
		seedHash,
	}

	if tx != nil {
		if _, err := tx.Exec(ctx, queryPlayer, args...); err != nil {
			return err
		}
		return s.ensurePlayerLedgerAccountTx(ctx, tx, actor.UserID, initialBalance)
	}

	if _, err := s.pool.Exec(ctx, queryPlayer, args...); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `
		insert into casino_accounts (user_id, type, code, balance, updated_at)
		values ($1, 'user', $2, $3, now())
		on conflict (code) do nothing
	`, actor.UserID, fmt.Sprintf("USER_%d", actor.UserID), initialBalance)
	return err
}

func (s *Store) ensurePlayerLedgerAccountTx(ctx context.Context, tx pgx.Tx, userID int64, balance int64) error {
	_, err := tx.Exec(ctx, `
		insert into casino_accounts (user_id, type, code, balance, updated_at)
		values ($1, 'user', $2, $3, now())
		on conflict (code) do nothing
	`, userID, fmt.Sprintf("USER_%d", userID), balance)
	return err
}

func (s *Store) ensurePlayerServerSeedTx(ctx context.Context, tx pgx.Tx, userID int64, serverSeed string) (string, string, error) {
	if strings.TrimSpace(serverSeed) != "" {
		return serverSeed, HashServerSeed(serverSeed), nil
	}

	serverSeed, err := GenerateServerSeed()
	if err != nil {
		return "", "", err
	}
	serverSeedHash := HashServerSeed(serverSeed)
	_, err = tx.Exec(ctx, `
		update casino_players
		set server_seed = $2,
			server_seed_hash = $3,
			updated_at = now()
		where user_id = $1
	`, userID, serverSeed, serverSeedHash)
	if err != nil {
		return "", "", err
	}
	return serverSeed, serverSeedHash, nil
}

func (s *Store) updatePlayerStatsTx(ctx context.Context, tx pgx.Tx, userID, wagered, won, gamesPlayed, rouletteRoundsPlayed int64) error {

	_, err := tx.Exec(ctx, `
		update casino_players
		set
			total_wagered          = total_wagered + $2,
			total_won              = total_won + $3,
			games_played           = games_played + $4,
			roulette_rounds_played = roulette_rounds_played + $5,
			last_game_at           = now(),
			level                  = (total_wagered + $2) / $6 + 1,
			xp_progress            = (total_wagered + $2) % $6,
			updated_at             = now()
		where user_id = $1
	`, userID, wagered, won, gamesPlayed, rouletteRoundsPlayed, levelStepWagered)
	return err
}

func computeLevel(totalWagered int64) (int64, int64) {
	if totalWagered < 0 {
		totalWagered = 0
	}
	return totalWagered/levelStepWagered + 1, totalWagered % levelStepWagered
}
