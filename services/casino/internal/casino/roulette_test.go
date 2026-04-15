package casino

import "testing"

func TestRoulettePayouts(t *testing.T) {
	cases := []struct {
		name          string
		bet           RouletteBet
		winningNumber int
		winningColor  string
		wantPayout    int64
		wantWon       bool
	}{
		{
			name:          "straight win",
			bet:           RouletteBet{BetType: "straight", BetValue: "7", Stake: 25},
			winningNumber: 7,
			winningColor:  "red",
			wantPayout:    900,
			wantWon:       true,
		},
		{
			name:          "green pays 36x",
			bet:           RouletteBet{BetType: "green", BetValue: "green", Stake: 10},
			winningNumber: 0,
			winningColor:  "green",
			wantPayout:    360,
			wantWon:       true,
		},
		{
			name:          "dozen1 pays 3x",
			bet:           RouletteBet{BetType: "dozen1", BetValue: "dozen1", Stake: 15},
			winningNumber: 12,
			winningColor:  "red",
			wantPayout:    45,
			wantWon:       true,
		},
		{
			name:          "dozen2 pays 3x",
			bet:           RouletteBet{BetType: "dozen2", BetValue: "dozen2", Stake: 15},
			winningNumber: 24,
			winningColor:  "black",
			wantPayout:    45,
			wantWon:       true,
		},
		{
			name:          "dozen3 pays 3x",
			bet:           RouletteBet{BetType: "dozen3", BetValue: "dozen3", Stake: 15},
			winningNumber: 36,
			winningColor:  "red",
			wantPayout:    45,
			wantWon:       true,
		},
		{
			name:          "red loses on green",
			bet:           RouletteBet{BetType: "red", BetValue: "red", Stake: 20},
			winningNumber: 0,
			winningColor:  "green",
			wantPayout:    0,
			wantWon:       false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotPayout, gotWon := roulettePayout(tc.bet, tc.winningNumber, tc.winningColor)
			if gotPayout != tc.wantPayout || gotWon != tc.wantWon {
				t.Fatalf("roulettePayout() = (%d, %v), want (%d, %v)", gotPayout, gotWon, tc.wantPayout, tc.wantWon)
			}
		})
	}
}

func TestNormalizeRouletteBet(t *testing.T) {
	t.Run("normalizes straight numbers", func(t *testing.T) {
		got, err := normalizeRouletteBet(RouletteBetInput{
			BetType:  " Straight ",
			BetValue: "07",
			Stake:    25,
		})
		if err != nil {
			t.Fatalf("normalizeRouletteBet() error = %v", err)
		}
		if got.BetType != "straight" || got.BetValue != "7" || got.Stake != 25 {
			t.Fatalf("unexpected normalized bet: %#v", got)
		}
	})

	t.Run("fills dozen bet value automatically", func(t *testing.T) {
		got, err := normalizeRouletteBet(RouletteBetInput{
			BetType: "DoZeN2",
			Stake:   50,
		})
		if err != nil {
			t.Fatalf("normalizeRouletteBet() error = %v", err)
		}
		if got.BetType != "dozen2" || got.BetValue != "dozen2" {
			t.Fatalf("unexpected normalized dozen bet: %#v", got)
		}
	})

	t.Run("rejects invalid stake", func(t *testing.T) {
		_, err := normalizeRouletteBet(RouletteBetInput{BetType: "red", Stake: 0})
		if err == nil {
			t.Fatal("expected stake validation error")
		}
	})

	t.Run("rejects unsupported bet type", func(t *testing.T) {
		_, err := normalizeRouletteBet(RouletteBetInput{BetType: "split", Stake: 10})
		if err == nil {
			t.Fatal("expected unsupported bet type error")
		}
	})
}
