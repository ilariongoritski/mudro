package casino

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *Store) GetCurrentRouletteState(ctx context.Context, userID int64) (RouletteState, error) {
	if cached, ok, err := s.loadRouletteSession(ctx); err == nil && ok {
		if userID > 0 {
			myBets, betsErr := s.getRouletteBetsForRound(ctx, cached.Round.ID, userID)
			if betsErr != nil && !errors.Is(betsErr, pgx.ErrNoRows) {
				return RouletteState{}, betsErr
			}
			cached.MyBets = myBets
		}
		return cached, nil
	} else if err != nil {
		return RouletteState{}, err
	}

	round, err := s.SyncRouletteRound(ctx)
	if err != nil {
		return RouletteState{}, err
	}

	history, err := s.GetRouletteHistory(ctx, 12)
	if err != nil {
		return RouletteState{}, err
	}

	myBets := make([]RouletteBet, 0)
	if userID > 0 {
		myBets, err = s.getRouletteBetsForRound(ctx, round.ID, userID)
		if err != nil {
			return RouletteState{}, err
		}
	}

	secondsLeft := int64(0)
	now := nowUTC()
	switch round.Status {
	case RoulettePhaseBetting:
		secondsLeft = maxInt64(0, int64(round.BettingClosesAt.Sub(now).Seconds()))
	case RoulettePhaseLocking:
		spinAt := round.BettingClosesAt.Add(RouletteLockDuration())
		secondsLeft = maxInt64(0, int64(spinAt.Sub(now).Seconds()))
	case RoulettePhaseSpinning:
		if round.SpinStartedAt != nil {
			resolveAt := round.SpinStartedAt.Add(RouletteSpinDuration())
			secondsLeft = maxInt64(0, int64(resolveAt.Sub(now).Seconds()))
		}
	case RoulettePhaseResult:
		if round.ResolvedAt != nil {
			nextAt := round.ResolvedAt.Add(RouletteResultDuration())
			secondsLeft = maxInt64(0, int64(nextAt.Sub(now).Seconds()))
		}
	}

	_ = s.cacheRouletteSession(ctx, round, history)

	return RouletteState{
		Round:         round,
		Phase:         round.Status,
		ServerTime:    now,
		SecondsLeft:   secondsLeft,
		RecentResults: history,
		MyBets:        myBets,
	}, nil
}

func (s *Store) SyncRouletteRound(ctx context.Context) (RouletteRound, error) {
	round, err := s.GetOrCreateRouletteRound(ctx)
	if err != nil {
		return RouletteRound{}, err
	}
	for i := 0; i < 4; i++ {
		next, err := s.TransitionRouletteRound(ctx, round)
		if err != nil {
			return RouletteRound{}, err
		}
		if next.ID == round.ID && next.Status == round.Status {
			history, historyErr := s.GetRouletteHistory(ctx, 12)
			if historyErr == nil {
				_ = s.cacheRouletteSession(ctx, next, history)
			}
			return next, nil
		}
		round = next
	}
	history, historyErr := s.GetRouletteHistory(ctx, 12)
	if historyErr == nil {
		_ = s.cacheRouletteSession(ctx, round, history)
	}
	return round, nil
}

func (s *Store) PlaceRouletteBets(ctx context.Context, actor ParticipantInput, roundID int64, bets []RouletteBetInput) (*RoulettePlaceBetsResponse, error) {
	if len(bets) == 0 {
		return nil, fmt.Errorf("at least one bet is required")
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

	if roundID == 0 {
		latest, err := s.getLatestRouletteRoundTx(ctx, tx)
		if err != nil {
			return nil, err
		}
		roundID = latest.ID
	}

	round, err := s.getRouletteRoundTx(ctx, tx, roundID, true)
	if err != nil {
		return nil, err
	}
	if round.Status != RoulettePhaseBetting || nowUTC().After(round.BettingClosesAt) {
		return nil, ErrRoundClosed
	}

	var totalStake int64
	normalized := make([]RouletteBetInput, 0, len(bets))
	for _, bet := range bets {
		normalizedBet, err := normalizeRouletteBet(bet)
		if err != nil {
			return nil, err
		}
		totalStake += normalizedBet.Stake
		normalized = append(normalized, normalizedBet)
	}

	var balance int64
	if err := tx.QueryRow(ctx, `select balance from casino_players where user_id = $1 for update`, actor.UserID).Scan(&balance); err != nil {
		return nil, err
	}
	if balance < totalStake {
		return nil, ErrInsufficientBalance
	}

	newBalance := balance - totalStake
	if err := s.setBalanceTx(ctx, tx, actor.UserID, newBalance); err != nil {
		return nil, err
	}

	placed := make([]RouletteBet, 0, len(normalized))
	for _, bet := range normalized {
		var placedBet RouletteBet
		err := tx.QueryRow(ctx, `
			insert into casino_roulette_bets (round_id, user_id, bet_type, bet_value, stake, payout_amount, status)
			values ($1, $2, $3, $4, $5, 0, 'placed')
			returning id, round_id, user_id, bet_type, bet_value, stake, payout_amount, status, created_at
		`, roundID, actor.UserID, bet.BetType, bet.BetValue, bet.Stake).Scan(
			&placedBet.ID,
			&placedBet.RoundID,
			&placedBet.UserID,
			&placedBet.BetType,
			&placedBet.BetValue,
			&placedBet.Stake,
			&placedBet.PayoutAmount,
			&placedBet.Status,
			&placedBet.CreatedAt,
		)
		if err != nil {
			if isUniqueViolation(err) {
				return nil, fmt.Errorf("duplicate roulette bet for %s:%s", bet.BetType, bet.BetValue)
			}
			return nil, err
		}
		placed = append(placed, placedBet)

		if err := s.recordTransferTx(ctx, tx, "bet_stake", actor.UserID, bet.Stake, map[string]any{"bet_id": placedBet.ID, "round_id": roundID}); err != nil {
			return nil, err
		}
	}

	if err := s.enqueueBalanceSyncTx(ctx, tx, actor.UserID, "roulette_bet", &newBalance); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true

	return &RoulettePlaceBetsResponse{
		RoundID: roundID,
		Balance: newBalance,
		Bets:    placed,
	}, nil
}

func (s *Store) InstantRouletteSpin(ctx context.Context, actor ParticipantInput, bets []RouletteBetInput) (*RouletteInstantSpinResponse, error) {
	if len(bets) == 0 {
		return nil, fmt.Errorf("at least one bet is required")
	}

	var totalStake int64
	normalized := make([]RouletteBetInput, 0, len(bets))
	for _, bet := range bets {
		norm, err := normalizeRouletteBet(bet)
		if err != nil {
			return nil, err
		}
		totalStake += norm.Stake
		normalized = append(normalized, norm)
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

	balance, _, clientSeed, serverSeed, nonce, _, err := s.getWalletStateForUpdate(ctx, tx, actor.UserID)
	if err != nil {
		return nil, err
	}
	serverSeed, serverSeedHash, err := s.ensurePlayerServerSeedTx(ctx, tx, actor.UserID, serverSeed)
	if err != nil {
		return nil, err
	}

	if balance < totalStake {
		return nil, ErrInsufficientBalance
	}

	winningNumber := drawRouletteNumber(NewFairness(serverSeed, clientSeed, int64(nonce)))
	winningColor := rouletteColor(winningNumber)
	displaySequence := buildRouletteDisplaySequence(winningNumber)
	resultSequence := buildRouletteResultSequence(displaySequence, winningNumber)

	now := nowUTC()
	var roundID int64
	err = tx.QueryRow(ctx, `
		insert into casino_roulette_rounds (
			status, winning_number, winning_color,
			display_sequence, result_sequence,
			betting_opens_at, betting_closes_at,
			spin_started_at, resolved_at,
			server_seed, server_seed_hash, client_seed, nonce
		) values (
			'result', $1, $2, $3::jsonb, $4::jsonb, $5, $5, $5, $5, $6, $7, $8, $9
		) returning id
	`, winningNumber, winningColor, marshalJSON(displaySequence), marshalJSON(resultSequence), now, serverSeed, serverSeedHash, clientSeed, nonce).Scan(&roundID)
	if err != nil {
		return nil, err
	}

	var totalPayout int64
	placedBets := make([]RouletteBet, 0, len(normalized))
	for _, b := range normalized {
		payout, won := roulettePayout(RouletteBet{BetType: b.BetType, BetValue: b.BetValue, Stake: b.Stake}, winningNumber, winningColor)
		totalPayout += payout

		status := "lost"
		if won {
			status = "win"
		}

		var pb RouletteBet
		err := tx.QueryRow(ctx, `
			insert into casino_roulette_bets (round_id, user_id, bet_type, bet_value, stake, payout_amount, status)
			values ($1, $2, $3, $4, $5, $6, $7)
			returning id, round_id, user_id, bet_type, bet_value, stake, payout_amount, status, created_at
		`, roundID, actor.UserID, b.BetType, b.BetValue, b.Stake, payout, status).Scan(
			&pb.ID, &pb.RoundID, &pb.UserID, &pb.BetType, &pb.BetValue, &pb.Stake, &pb.PayoutAmount, &pb.Status, &pb.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		placedBets = append(placedBets, pb)

		if err := s.recordTransferTx(ctx, tx, "bet_stake", actor.UserID, b.Stake, map[string]any{"bet_id": pb.ID, "game": "roulette_instant"}); err != nil {
			return nil, err
		}

		if payout > 0 {
			if err := s.recordTransferTx(ctx, tx, "bet_payout", actor.UserID, payout, map[string]any{"bet_id": pb.ID, "game": "roulette_instant"}); err != nil {
				return nil, err
			}
		}
	}

	newBalance := balance - totalStake + totalPayout
	if err := s.setWalletStateTx(ctx, tx, actor.UserID, newBalance, nil); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `update casino_players set current_nonce = current_nonce + 1 where user_id = $1`, actor.UserID); err != nil {
		return nil, err
	}

	if err := s.updatePlayerStatsTx(ctx, tx, actor.UserID, totalStake, totalPayout, 1, 1); err != nil {
		return nil, err
	}

	netResult := totalPayout - totalStake
	activityStatus := "LOST"
	if totalPayout > totalStake {
		activityStatus = "WIN"
	} else if totalPayout == totalStake {
		activityStatus = "PUSH"
	}

	if err := s.insertActivityTx(ctx, tx, actor.UserID, "roulette_instant", fmt.Sprintf("%d", roundID), totalStake, totalPayout, netResult, activityStatus, map[string]any{
		"winning_number":   winningNumber,
		"winning_color":    winningColor,
		"result_sequence":  resultSequence,
		"client_seed":      clientSeed,
		"nonce":            nonce,
		"server_seed_hash": serverSeedHash,
	}); err != nil {
		return nil, err
	}

	if err := s.enqueueBalanceSyncTx(ctx, tx, actor.UserID, "roulette_instant", &newBalance); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true

	return &RouletteInstantSpinResponse{
		WinningNumber:   winningNumber,
		WinningColor:    winningColor,
		DisplaySequence: displaySequence,
		ResultSequence:  resultSequence,
		PayoutAmount:    totalPayout,
		Balance:         newBalance,
		Bets:            placedBets,
	}, nil
}

func (s *Store) GetRouletteHistory(ctx context.Context, limit int) ([]RouletteHistoryItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.pool.Query(ctx, `
		select id, winning_number, winning_color, resolved_at
		from casino_roulette_rounds
		where status = 'result' and winning_number is not null and resolved_at is not null
		order by resolved_at desc, id desc
		limit $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]RouletteHistoryItem, 0, limit)
	for rows.Next() {
		var item RouletteHistoryItem
		if err := rows.Scan(&item.RoundID, &item.WinningNumber, &item.WinningColor, &item.ResolvedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetOrCreateRouletteRound(ctx context.Context) (RouletteRound, error) {
	round, err := s.getLatestRouletteRound(ctx)
	if err == nil {
		return round, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return RouletteRound{}, err
	}
	return s.createRouletteRound(ctx, nowUTC())
}

func (s *Store) TransitionRouletteRound(ctx context.Context, round RouletteRound) (RouletteRound, error) {
	now := nowUTC()
	switch round.Status {
	case RoulettePhaseBetting:
		if !now.Before(round.BettingClosesAt) {
			return s.updateRouletteRoundStatus(ctx, round.ID, RoulettePhaseLocking, nil, "", nil, nil, nil, nil)
		}
	case RoulettePhaseLocking:
		spinAt := round.BettingClosesAt.Add(RouletteLockDuration())
		if !now.Before(spinAt) {
			fairness := NewFairness(round.ServerSeed, round.ClientSeed, round.Nonce)
			winningNumber := drawRouletteNumber(fairness)
			winningColor := rouletteColor(winningNumber)
			displaySequence := buildRouletteDisplaySequence(winningNumber)
			resultSequence := buildRouletteResultSequence(displaySequence, winningNumber)
			spinStartedAt := now
			return s.updateRouletteRoundStatus(ctx, round.ID, RoulettePhaseSpinning, &winningNumber, winningColor, &spinStartedAt, nil, displaySequence, resultSequence)
		}
	case RoulettePhaseSpinning:
		if round.SpinStartedAt != nil {
			resolveAt := round.SpinStartedAt.Add(RouletteSpinDuration())
			if !now.Before(resolveAt) {
				resolvedAt := now
				updated, err := s.updateRouletteRoundStatus(ctx, round.ID, RoulettePhaseResult, round.WinningNumber, round.WinningColor, round.SpinStartedAt, &resolvedAt, round.DisplaySequence, round.ResultSequence)
				if err != nil {
					return RouletteRound{}, err
				}
				if err := s.settleRouletteRound(ctx, updated); err != nil {
					return RouletteRound{}, err
				}
				return updated, nil
			}
		}
	case RoulettePhaseResult:
		if round.ResolvedAt != nil {
			nextAt := round.ResolvedAt.Add(RouletteResultDuration())
			if !now.Before(nextAt) {
				return s.createRouletteRound(ctx, now)
			}
		}
	}
	return round, nil
}

func (s *Store) getLatestRouletteRound(ctx context.Context) (RouletteRound, error) {
	return s.getRouletteRoundQuery(ctx, s.pool)
}

func (s *Store) getLatestRouletteRoundTx(ctx context.Context, tx pgx.Tx) (RouletteRound, error) {
	return s.getRouletteRoundQuery(ctx, tx)
}

func (s *Store) getRouletteRoundQuery(ctx context.Context, q queryable) (RouletteRound, error) {
	return scanRouletteRound(q.QueryRow(ctx, `
		select
			id,
			status,
			winning_number,
			winning_color,
			display_sequence::text,
			result_sequence::text,
			betting_opens_at,
			betting_closes_at,
			spin_started_at,
			resolved_at,
			created_at,
			server_seed,
			server_seed_hash,
			client_seed,
			nonce,
			round_hash
		from casino_roulette_rounds
		order by id desc
		limit 1
	`))
}

func (s *Store) getRouletteRoundTx(ctx context.Context, tx pgx.Tx, roundID int64, forUpdate bool) (RouletteRound, error) {
	query := `
		select
			id,
			status,
			winning_number,
			winning_color,
			display_sequence::text,
			result_sequence::text,
			betting_opens_at,
			betting_closes_at,
			spin_started_at,
			resolved_at,
			created_at,
			server_seed,
			server_seed_hash,
			client_seed,
			nonce,
			round_hash
		from casino_roulette_rounds
		where id = $1
	`
	if forUpdate {
		query += ` for update`
	}
	return scanRouletteRound(tx.QueryRow(ctx, query, roundID))
}

func (s *Store) createRouletteRound(ctx context.Context, now time.Time) (RouletteRound, error) {
	serverSeed, err := GenerateServerSeed()
	if err != nil {
		return RouletteRound{}, err
	}
	serverSeedHash := HashServerSeed(serverSeed)

	return scanRouletteRound(s.pool.QueryRow(ctx, `
		insert into casino_roulette_rounds (
			status,
			display_sequence,
			result_sequence,
			betting_opens_at,
			betting_closes_at,
			server_seed,
			server_seed_hash
		)
		values ('betting', '[]'::jsonb, '[]'::jsonb, $1, $2, $3, $4)
		returning
			id,
			status,
			winning_number,
			winning_color,
			display_sequence::text,
			result_sequence::text,
			betting_opens_at,
			betting_closes_at,
			spin_started_at,
			resolved_at,
			created_at,
			server_seed,
			server_seed_hash,
			client_seed,
			nonce,
			round_hash
	`, now, now.Add(RouletteBettingDuration()), serverSeed, serverSeedHash))
}

func (s *Store) updateRouletteRoundStatus(
	ctx context.Context,
	roundID int64,
	status RoulettePhase,
	winningNumber *int,
	winningColor string,
	spinStartedAt *time.Time,
	resolvedAt *time.Time,
	displaySequence []int,
	resultSequence []int,
) (RouletteRound, error) {
	return scanRouletteRound(s.pool.QueryRow(ctx, `
		update casino_roulette_rounds
		set status = $2,
			winning_number = $3,
			winning_color = $4,
			display_sequence = $5::jsonb,
			result_sequence = $6::jsonb,
			spin_started_at = $7,
			resolved_at = $8
		where id = $1
		returning
			id,
			status,
			winning_number,
			winning_color,
			display_sequence::text,
			result_sequence::text,
			betting_opens_at,
			betting_closes_at,
			spin_started_at,
			resolved_at,
			created_at,
			server_seed,
			server_seed_hash,
			client_seed,
			nonce,
			round_hash
	`, roundID, status, winningNumber, nullableString(winningColor), marshalJSON(displaySequence), marshalJSON(resultSequence), spinStartedAt, resolvedAt))
}

// GetRouletteBetsForRound returns all bets placed by a user in the given round.
// Used by the SSE handler to attach per-user bet data to the shared hub state.
func (s *Store) GetRouletteBetsForRound(ctx context.Context, roundID, userID int64) ([]RouletteBet, error) {
	return s.getRouletteBetsForRound(ctx, roundID, userID)
}

func (s *Store) getRouletteBetsForRound(ctx context.Context, roundID, userID int64) ([]RouletteBet, error) {
	rows, err := s.pool.Query(ctx, `
		select id, round_id, user_id, bet_type, bet_value, stake, payout_amount, status, created_at
		from casino_roulette_bets
		where round_id = $1 and user_id = $2
		order by created_at asc, id asc
	`, roundID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]RouletteBet, 0)
	for rows.Next() {
		var item RouletteBet
		if err := rows.Scan(&item.ID, &item.RoundID, &item.UserID, &item.BetType, &item.BetValue, &item.Stake, &item.PayoutAmount, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) settleRouletteRound(ctx context.Context, round RouletteRound) error {
	if round.WinningNumber == nil {
		return fmt.Errorf("roulette round %d has no winning number", round.ID)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	currentRound, err := s.getRouletteRoundTx(ctx, tx, round.ID, true)
	if err != nil {
		return err
	}
	if currentRound.Status != RoulettePhaseResult {
		return fmt.Errorf("roulette round %d is not ready for settlement", round.ID)
	}

	rows, err := tx.Query(ctx, `
		select id, round_id, user_id, bet_type, bet_value, stake, payout_amount, status, created_at
		from casino_roulette_bets
		where round_id = $1 and status = 'placed'
		order by created_at asc, id asc
	`, round.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type playerDelta struct {
		wagered        int64
		won            int64
		rouletteRounds int64
		gamesPlayed    int64
		balanceDelta   int64
	}

	playerDeltas := make(map[int64]*playerDelta)

	for rows.Next() {
		var bet RouletteBet
		if err := rows.Scan(&bet.ID, &bet.RoundID, &bet.UserID, &bet.BetType, &bet.BetValue, &bet.Stake, &bet.PayoutAmount, &bet.Status, &bet.CreatedAt); err != nil {
			return err
		}
		payout, won := roulettePayout(bet, *round.WinningNumber, round.WinningColor)
		status := "LOST"
		if won {
			status = "WIN"
		}

		if _, err := tx.Exec(ctx, `
			update casino_roulette_bets
			set payout_amount = $2,
				status = $3
			where id = $1
		`, bet.ID, payout, strings.ToLower(status)); err != nil {
			return err
		}

		if payout > 0 {
			if err := s.recordTransferTx(ctx, tx, "bet_payout", bet.UserID, payout, map[string]any{"bet_id": bet.ID, "round_id": round.ID}); err != nil {
				return err
			}
		}

		delta := playerDeltas[bet.UserID]
		if delta == nil {
			delta = &playerDelta{
				rouletteRounds: 1,
				gamesPlayed:    1,
			}
			playerDeltas[bet.UserID] = delta
		}
		delta.wagered += bet.Stake
		delta.won += payout
		delta.balanceDelta += payout

		if err := s.insertActivityTx(ctx, tx, bet.UserID, "roulette", fmt.Sprintf("%d:%d", round.ID, bet.ID), bet.Stake, payout, payout-bet.Stake, status, map[string]any{
			"round_id":         round.ID,
			"winning_number":   round.WinningNumber,
			"winning_color":    round.WinningColor,
			"bet_type":         bet.BetType,
			"bet_value":        bet.BetValue,
			"result_sequence":  round.ResultSequence,
			"client_seed":      round.ClientSeed,
			"nonce":            round.Nonce,
			"server_seed_hash": round.ServerSeedHash,
		}); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for userID, delta := range playerDeltas {
		var balance int64
		if err := tx.QueryRow(ctx, `select balance from casino_players where user_id = $1 for update`, userID).Scan(&balance); err != nil {
			return err
		}
		if err := s.setBalanceTx(ctx, tx, userID, balance+delta.balanceDelta); err != nil {
			return err
		}
		if err := s.updatePlayerStatsTx(ctx, tx, userID, delta.wagered, delta.won, delta.gamesPlayed, delta.rouletteRounds); err != nil {
			return err
		}
		targetBalance := balance + delta.balanceDelta
		if err := s.enqueueBalanceSyncTx(ctx, tx, userID, "roulette_settlement", &targetBalance); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true

	history, historyErr := s.GetRouletteHistory(ctx, 12)
	if historyErr == nil {
		_ = s.cacheRouletteSession(ctx, round, history)
	}
	return nil
}

func scanRouletteRound(row pgx.Row) (RouletteRound, error) {
	var (
		item              RouletteRound
		displaySequenceTx string
		resultSequenceTx  string
		winningNumber     *int
		winningColor      *string
	)
	if err := row.Scan(
		&item.ID,
		&item.Status,
		&winningNumber,
		&winningColor,
		&displaySequenceTx,
		&resultSequenceTx,
		&item.BettingOpensAt,
		&item.BettingClosesAt,
		&item.SpinStartedAt,
		&item.ResolvedAt,
		&item.CreatedAt,
		&item.ServerSeed,
		&item.ServerSeedHash,
		&item.ClientSeed,
		&item.Nonce,
		&item.RoundHash,
	); err != nil {
		return RouletteRound{}, err
	}
	item.WinningNumber = winningNumber
	if winningColor != nil {
		item.WinningColor = *winningColor
	}
	if err := json.Unmarshal([]byte(displaySequenceTx), &item.DisplaySequence); err != nil {
		return RouletteRound{}, err
	}
	if err := json.Unmarshal([]byte(resultSequenceTx), &item.ResultSequence); err != nil {
		return RouletteRound{}, err
	}
	return item, nil
}

func normalizeRouletteBet(bet RouletteBetInput) (RouletteBetInput, error) {
	bet.BetType = strings.TrimSpace(strings.ToLower(bet.BetType))
	bet.BetValue = strings.TrimSpace(strings.ToLower(bet.BetValue))
	if bet.Stake <= 0 {
		return RouletteBetInput{}, fmt.Errorf("stake must be positive")
	}
	switch bet.BetType {
	case "straight":
		number, err := parseRouletteNumber(bet.BetValue)
		if err != nil {
			return RouletteBetInput{}, err
		}
		bet.BetValue = fmt.Sprintf("%d", number)
	case "red", "black", "green", "odd", "even", "low", "high", "dozen1", "dozen2", "dozen3":
		if bet.BetValue == "" {
			bet.BetValue = bet.BetType
		}
	default:
		return RouletteBetInput{}, fmt.Errorf("unsupported bet type %q", bet.BetType)
	}
	return bet, nil
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func nullableString(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}
