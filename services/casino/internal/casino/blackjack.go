package casino

import ()

type BlackjackEngine struct{}

func NewBlackjackEngine() *BlackjackEngine {
	return &BlackjackEngine{}
}

func (e *BlackjackEngine) NewGame(bet int64) (*BlackjackState, error) {
	deck := e.generateDeck()
	playerHand := BlackjackHand{Cards: []BlackjackCard{}}
	dealerHand := BlackjackHand{Cards: []BlackjackCard{}}

	// Deal initial cards
	for i := 0; i < 2; i++ {
		playerHand.Cards = append(playerHand.Cards, e.drawCard(&deck))
		dealerHand.Cards = append(dealerHand.Cards, e.drawCard(&deck))
	}

	playerHand.Score = e.calculateScore(playerHand.Cards)
	dealerHand.Score = e.calculateScore(dealerHand.Cards)

	state := &BlackjackState{
		Bet:        bet,
		PlayerHand: playerHand,
		DealerHand: dealerHand,
		Status:     BlackjackStatusPlayerTurn,
		CreatedAt:  nowUTC(),
	}

	// Check for immediate blackjack
	if playerHand.Score == 21 {
		return e.Resolve(state)
	}

	return state, nil
}

func (e *BlackjackEngine) Hit(state *BlackjackState) (*BlackjackState, error) {
	if state.Status != BlackjackStatusPlayerTurn {
		return state, nil
	}

	deck := e.generateDeckFromHands(state.PlayerHand, state.DealerHand)
	state.PlayerHand.Cards = append(state.PlayerHand.Cards, e.drawCard(&deck))
	state.PlayerHand.Score = e.calculateScore(state.PlayerHand.Cards)

	if state.PlayerHand.Score > 21 {
		state.PlayerHand.IsBust = true
		return e.Resolve(state)
	}

	if state.PlayerHand.Score == 21 {
		return e.Resolve(state)
	}

	return state, nil
}

func (e *BlackjackEngine) Stand(state *BlackjackState) (*BlackjackState, error) {
	if state.Status != BlackjackStatusPlayerTurn {
		return state, nil
	}
	state.Status = BlackjackStatusDealerTurn
	return e.Resolve(state)
}

func (e *BlackjackEngine) Resolve(state *BlackjackState) (*BlackjackState, error) {
	if state.PlayerHand.IsBust {
		state.Status = BlackjackStatusResolved
		state.Winner = "dealer"
		state.Payout = 0
		return state, nil
	}

	// Dealer Turn AI
	deck := e.generateDeckFromHands(state.PlayerHand, state.DealerHand)
	for state.DealerHand.Score < 17 {
		state.DealerHand.Cards = append(state.DealerHand.Cards, e.drawCard(&deck))
		state.DealerHand.Score = e.calculateScore(state.DealerHand.Cards)
	}

	if state.DealerHand.Score > 21 {
		state.DealerHand.IsBust = true
	}

	state.Status = BlackjackStatusResolved
	if state.DealerHand.IsBust || state.PlayerHand.Score > state.DealerHand.Score {
		state.Winner = "player"
		state.Payout = state.Bet * 2
		if state.PlayerHand.Score == 21 && len(state.PlayerHand.Cards) == 2 {
			state.Payout = int64(float64(state.Bet) * 2.5) // Blackjack payout 3:2
		}
	} else if state.PlayerHand.Score < state.DealerHand.Score {
		state.Winner = "dealer"
		state.Payout = 0
	} else {
		state.Winner = "push"
		state.Payout = state.Bet
	}

	return state, nil
}

func (e *BlackjackEngine) generateDeck() []BlackjackCard {
	suits := []string{"hearts", "diamonds", "clubs", "spades"}
	ranks := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	values := map[string]int{
		"2": 2, "3": 3, "4": 4, "5": 5, "6": 6, "7": 7, "8": 8, "9": 9, "10": 10,
		"J": 10, "Q": 10, "K": 10, "A": 11,
	}

	deck := make([]BlackjackCard, 0, 52)
	for _, suit := range suits {
		for _, rank := range ranks {
			deck = append(deck, BlackjackCard{Suit: suit, Rank: rank, Value: values[rank]})
		}
	}
	return deck
}

func (e *BlackjackEngine) generateDeckFromHands(player, dealer BlackjackHand) []BlackjackCard {
	full := e.generateDeck()
	used := make(map[string]bool)
	for _, c := range player.Cards {
		used[c.Suit+c.Rank] = true
	}
	for _, c := range dealer.Cards {
		used[c.Suit+c.Rank] = true
	}

	deck := make([]BlackjackCard, 0)
	for _, c := range full {
		if !used[c.Suit+c.Rank] {
			deck = append(deck, c)
		}
	}
	return deck
}

func (e *BlackjackEngine) drawCard(deck *[]BlackjackCard) BlackjackCard {
	if len(*deck) == 0 {
		return BlackjackCard{}
	}
	idx := DrawInt(len(*deck))
	card := (*deck)[idx]
	*deck = append((*deck)[:idx], (*deck)[idx+1:]...)
	return card
}

func (e *BlackjackEngine) calculateScore(cards []BlackjackCard) int {
	score := 0
	aces := 0
	for _, card := range cards {
		score += card.Value
		if card.Rank == "A" {
			aces++
		}
	}
	for score > 21 && aces > 0 {
		score -= 10
		aces--
	}
	return score
}
