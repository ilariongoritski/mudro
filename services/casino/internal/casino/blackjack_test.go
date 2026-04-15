package casino

import "testing"

func TestBlackjackResolvePlayerBust(t *testing.T) {
	engine := NewBlackjackEngine()
	state := &BlackjackState{
		Bet: 100,
		PlayerHand: BlackjackHand{
			Cards: []BlackjackCard{
				{Rank: "10", Value: 10},
				{Rank: "9", Value: 9},
				{Rank: "5", Value: 5},
			},
			Score:  24,
			IsBust: true,
		},
		DealerHand: BlackjackHand{
			Cards: []BlackjackCard{
				{Rank: "10", Value: 10},
				{Rank: "7", Value: 7},
			},
			Score: 17,
		},
		Status: BlackjackStatusPlayerTurn,
	}

	got, err := engine.Resolve(state, nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.Status != BlackjackStatusResolved || got.Winner != "dealer" || got.Payout != 0 {
		t.Fatalf("unexpected resolved state: %#v", got)
	}
}

func TestBlackjackResolveNaturalBlackjackPaysThreeToTwo(t *testing.T) {
	engine := NewBlackjackEngine()
	state := &BlackjackState{
		Bet: 100,
		PlayerHand: BlackjackHand{
			Cards: []BlackjackCard{
				{Rank: "A", Value: 11},
				{Rank: "K", Value: 10},
			},
			Score: 21,
		},
		DealerHand: BlackjackHand{
			Cards: []BlackjackCard{
				{Rank: "10", Value: 10},
				{Rank: "7", Value: 7},
			},
			Score: 17,
		},
		Status: BlackjackStatusDealerTurn,
	}

	got, err := engine.Resolve(state, nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.Winner != "player" || got.Payout != 250 {
		t.Fatalf("unexpected payout: winner=%q payout=%d", got.Winner, got.Payout)
	}
}

func TestBlackjackResolvePush(t *testing.T) {
	engine := NewBlackjackEngine()
	state := &BlackjackState{
		Bet: 80,
		PlayerHand: BlackjackHand{
			Cards: []BlackjackCard{
				{Rank: "10", Value: 10},
				{Rank: "Q", Value: 10},
			},
			Score: 20,
		},
		DealerHand: BlackjackHand{
			Cards: []BlackjackCard{
				{Rank: "K", Value: 10},
				{Rank: "10", Value: 10},
			},
			Score: 20,
		},
		Status: BlackjackStatusDealerTurn,
	}

	got, err := engine.Resolve(state, nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.Winner != "push" || got.Payout != 80 {
		t.Fatalf("unexpected push result: winner=%q payout=%d", got.Winner, got.Payout)
	}
}

func TestBlackjackStandResolvesWhenDealerAlreadyReady(t *testing.T) {
	engine := NewBlackjackEngine()
	state := &BlackjackState{
		Bet: 60,
		PlayerHand: BlackjackHand{
			Cards: []BlackjackCard{
				{Rank: "10", Value: 10},
				{Rank: "8", Value: 8},
			},
			Score: 18,
		},
		DealerHand: BlackjackHand{
			Cards: []BlackjackCard{
				{Rank: "10", Value: 10},
				{Rank: "7", Value: 7},
			},
			Score: 17,
		},
		Status: BlackjackStatusPlayerTurn,
	}

	got, err := engine.Stand(state, nil)
	if err != nil {
		t.Fatalf("Stand() error = %v", err)
	}
	if got.Status != BlackjackStatusResolved || got.Winner != "player" || got.Payout != 120 {
		t.Fatalf("unexpected stand result: %#v", got)
	}
}
