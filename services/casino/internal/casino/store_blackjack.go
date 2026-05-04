package casino

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (s *Store) BlackjackGetState(ctx context.Context, userID int64) (*BlackjackState, error) {
	var (
		state     BlackjackState
		pHandJson []byte
		dHandJson []byte
	)
	err := s.pool.QueryRow(ctx, `
		select id, user_id, bet, player_hand, dealer_hand, status, winner, payout, created_at, server_seed, client_seed, nonce
		from casino_blackjack_games
		where user_id = $1 and status != 'resolved'
		order by created_at desc
		limit 1
	`, userID).Scan(
		&state.ID,
		&state.UserID,
		&state.Bet,
		&pHandJson,
		&dHandJson,
		&state.Status,
		&state.Winner,
		&state.Payout,
		&state.CreatedAt,
		&state.ServerSeed,
		&state.ClientSeed,
		&state.Nonce,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(pHandJson, &state.PlayerHand); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(dHandJson, &state.DealerHand); err != nil {
		return nil, err
	}
	return &state, nil
}

func (s *Store) BlackjackStart(ctx context.Context, actor ParticipantInput, bet int64) (*BlackjackState, error) {
	if bet <= 0 {
		return nil, errors.New("bet must be positive")
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

	var activeID int64
	err = tx.QueryRow(ctx, `select id from casino_blackjack_games where user_id = $1 and status != 'resolved'`, actor.UserID).Scan(&activeID)
	if err == nil {
		return nil, errors.New("already have an active blackjack game")
	}

	balance, _, clientSeed, serverSeed, nonce, _, err := s.getWalletStateForUpdate(ctx, tx, actor.UserID)
	if err != nil {
		return nil, err
	}
	serverSeed, serverSeedHash, err := s.ensurePlayerServerSeedTx(ctx, tx, actor.UserID, serverSeed)
	if err != nil {
		return nil, err
	}

	if balance < bet {
		return nil, ErrInsufficientBalance
	}

	state, err := s.blackjack.NewGame(bet, NewFairness(serverSeed, clientSeed, int64(nonce)))
	if err != nil {
		return nil, err
	}
	state.UserID = actor.UserID
	state.ServerSeed = serverSeed
	state.ClientSeed = clientSeed
	state.Nonce = int64(nonce)

	newBalance := balance - bet
	if err := s.setBalanceTx(ctx, tx, actor.UserID, newBalance); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `update casino_players set current_nonce = current_nonce + 1 where user_id = $1`, actor.UserID); err != nil {
		return nil, err
	}

	if err := s.recordTransferTx(ctx, tx, "bet_stake", actor.UserID, bet, map[string]any{"game": "blackjack"}); err != nil {
		return nil, err
	}

	err = tx.QueryRow(ctx, `
		insert into casino_blackjack_games (
			user_id, bet, player_hand, dealer_hand, status, winner, payout, created_at,
			server_seed, server_seed_hash, client_seed, nonce
		)
		values ($1, $2, $3::jsonb, $4::jsonb, $5, $6, $7, $8, $9, $10, $11, $12)
		returning id
	`, state.UserID, state.Bet, marshalJSON(state.PlayerHand), marshalJSON(state.DealerHand), state.Status, state.Winner, state.Payout, state.CreatedAt, serverSeed, serverSeedHash, clientSeed, nonce).Scan(&state.ID)
	if err != nil {
		return nil, err
	}

	if state.Status == BlackjackStatusResolved {
		if err := s.resolveBlackjackTx(ctx, tx, state); err != nil {
			return nil, err
		}
	} else {
		if err := s.enqueueBalanceSyncTx(ctx, tx, actor.UserID, "blackjack_start", &newBalance); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true
	return state, nil
}

func (s *Store) BlackjackAction(ctx context.Context, actor ParticipantInput, action BlackjackAction) (*BlackjackState, error) {
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

	state, err := s.blackjackGetStateTx(ctx, tx, actor.UserID)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, errors.New("no active game")
	}

	var next *BlackjackState
	fairness := NewFairness(state.ServerSeed, state.ClientSeed, state.Nonce)
	switch action {
	case BlackjackActionHit:
		next, err = s.blackjack.Hit(state, fairness)
	case BlackjackActionStand:
		next, err = s.blackjack.Stand(state, fairness)
	default:
		return nil, errors.New("invalid action")
	}
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		update casino_blackjack_games
		set player_hand = $1::jsonb, dealer_hand = $2::jsonb, status = $3, winner = $4, payout = $5
		where id = $6
	`, marshalJSON(next.PlayerHand), marshalJSON(next.DealerHand), next.Status, next.Winner, next.Payout, next.ID)
	if err != nil {
		return nil, err
	}

	if next.Status == BlackjackStatusResolved {
		if err := s.resolveBlackjackTx(ctx, tx, next); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	committed = true
	return next, nil
}

func (s *Store) blackjackGetStateTx(ctx context.Context, tx pgx.Tx, userID int64) (*BlackjackState, error) {
	var (
		state     BlackjackState
		pHandJson []byte
		dHandJson []byte
	)
	err := tx.QueryRow(ctx, `
		select id, user_id, bet, player_hand, dealer_hand, status, winner, payout, created_at, server_seed, client_seed, nonce
		from casino_blackjack_games
		where user_id = $1 and status != 'resolved'
		for update
	`, userID).Scan(
		&state.ID,
		&state.UserID,
		&state.Bet,
		&pHandJson,
		&dHandJson,
		&state.Status,
		&state.Winner,
		&state.Payout,
		&state.CreatedAt,
		&state.ServerSeed,
		&state.ClientSeed,
		&state.Nonce,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(pHandJson, &state.PlayerHand); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(dHandJson, &state.DealerHand); err != nil {
		return nil, err
	}
	return &state, nil
}

func (s *Store) resolveBlackjackTx(ctx context.Context, tx pgx.Tx, state *BlackjackState) error {
	balance, _, _, _, _, _, err := s.getWalletStateForUpdate(ctx, tx, state.UserID)
	if err != nil {
		return err
	}
	newBalance := balance + state.Payout
	if err := s.setBalanceTx(ctx, tx, state.UserID, newBalance); err != nil {
		return err
	}

	if state.Payout > 0 {
		if err := s.recordTransferTx(ctx, tx, "bet_payout", state.UserID, state.Payout, map[string]any{"game": "blackjack", "game_id": state.ID}); err != nil {
			return err
		}
	}

	if err := s.updatePlayerStatsTx(ctx, tx, state.UserID, state.Bet, state.Payout, 1, 0); err != nil {
		return err
	}

	status := strings.ToUpper(state.Winner)
	if status == "" {
		status = "FINISHED"
	}

	if err := s.insertActivityTx(ctx, tx, state.UserID, "blackjack", fmt.Sprintf("%d", state.ID), state.Bet, state.Payout, state.Payout-state.Bet, status, map[string]any{
		"player_hand":      state.PlayerHand,
		"dealer_hand":      state.DealerHand,
		"winner":           state.Winner,
		"client_seed":      state.ClientSeed,
		"nonce":            state.Nonce,
		"server_seed_hash": HashServerSeed(state.ServerSeed),
	}); err != nil {
		return err
	}

	return s.enqueueBalanceSyncTx(ctx, tx, state.UserID, "blackjack_resolve", &newBalance)
}
