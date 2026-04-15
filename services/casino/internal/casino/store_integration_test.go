package casino

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testCasinoIntegrationStore(t *testing.T) *Store {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv("MUDRO_CASINO_INTEGRATION_TEST_DSN"))
	if dsn == "" {
		t.Skip("skip integration test: set MUDRO_CASINO_INTEGRATION_TEST_DSN to an isolated migrated casino database")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("skip integration test: db connect: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("skip integration test: db ping: %v", err)
	}

	if _, err := pool.Exec(ctx, `
		truncate table
			casino_activity_reactions,
			casino_game_activity,
			casino_bonus_claims,
			casino_roulette_sessions,
			casino_roulette_bets,
			casino_roulette_rounds,
			casino_blackjack_games,
			casino_spins,
			casino_balance_sync_queue,
			casino_ledger_entries,
			casino_transfers,
			casino_accounts,
			casino_players
		restart identity cascade;
		delete from casino_config;
	`); err != nil {
		t.Skipf("skip integration test: truncate casino schema: %v", err)
	}

	if _, err := pool.Exec(ctx, `
		insert into casino_accounts (type, code, balance)
		values
			('system', 'SYSTEM_HOUSE_POOL', 1000000),
			('user', 'USER_101', 500)
		on conflict (code) do nothing
	`); err != nil {
		t.Skipf("skip integration test: seed ledger accounts: %v", err)
	}

	return NewStore(pool, NewEngine())
}

func TestCasinoSmokeIntegration(t *testing.T) {
	store := testCasinoIntegrationStore(t)
	ctx := context.Background()
	actor := ParticipantInput{
		UserID:   101,
		Username: "smoke",
	}

	balance, freeSpins, bonusClaimed, err := store.GetBalanceDetails(ctx, actor)
	if err != nil {
		t.Fatalf("GetBalanceDetails() error = %v", err)
	}
	if balance <= 0 || freeSpins != 0 || bonusClaimed {
		t.Fatalf("unexpected initial wallet state: balance=%d freeSpins=%d bonusClaimed=%v", balance, freeSpins, bonusClaimed)
	}

	spinResult, err := store.Spin(ctx, actor, 10)
	if err != nil {
		t.Fatalf("Spin() error = %v", err)
	}
	if len(spinResult.Symbols) != 3 {
		t.Fatalf("spin symbols len = %d, want 3", len(spinResult.Symbols))
	}

	rouletteResult, err := store.InstantRouletteSpin(ctx, actor, []RouletteBetInput{{
		BetType: "red",
		Stake:   10,
	}})
	if err != nil {
		t.Fatalf("InstantRouletteSpin() error = %v", err)
	}
	if len(rouletteResult.Bets) != 1 {
		t.Fatalf("roulette bets len = %d, want 1", len(rouletteResult.Bets))
	}

	plinkoResult, err := store.DropPlinko(ctx, actor, PlinkoDropRequest{
		Bet:  10,
		Risk: PlinkoRiskMedium,
	})
	if err != nil {
		t.Fatalf("DropPlinko() error = %v", err)
	}
	if plinkoResult.Rows <= 0 || len(plinkoResult.Path) != plinkoResult.Rows {
		t.Fatalf("unexpected plinko result: %#v", plinkoResult)
	}

	game, err := store.BlackjackStart(ctx, actor, 10)
	if err != nil {
		t.Fatalf("BlackjackStart() error = %v", err)
	}
	finalState := game
	if game.Status != BlackjackStatusResolved {
		finalState, err = store.BlackjackAction(ctx, actor, BlackjackActionStand)
		if err != nil {
			t.Fatalf("BlackjackAction() error = %v", err)
		}
	}
	if finalState.Status != BlackjackStatusResolved {
		t.Fatalf("blackjack status = %q, want %q", finalState.Status, BlackjackStatusResolved)
	}

	history, err := store.GetHistory(ctx, actor.UserID, 10)
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}
	if len(history) == 0 {
		t.Fatal("expected slot history after smoke flow")
	}

	activity, err := store.GetActivity(ctx, actor.UserID, 20)
	if err != nil {
		t.Fatalf("GetActivity() error = %v", err)
	}
	if len(activity) < 4 {
		t.Fatalf("activity len = %d, want at least 4", len(activity))
	}
}
