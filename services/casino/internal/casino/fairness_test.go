package casino

import "testing"

func TestFairnessDeterministicRolls(t *testing.T) {
	cfg := Config{
		RTPPercent:     95.0,
		InitialBalance: 1000,
		SymbolWeights:  map[string]int{"X": 50, "Y": 50},
		Paytable:       map[string]int64{"X": 2, "Y": 2},
	}

	eng := NewEngine()
	eng.EnableFairness("serverseed123", "clientseedABC", 42)

	firstSymbols, firstWin, err := eng.Spin(cfg, 100)
	if err != nil {
		t.Fatalf("unexpected error on first spin: %v", err)
	}
	if len(firstSymbols) != 3 {
		t.Fatalf("expected 3 symbols, got %d", len(firstSymbols))
	}

	// Spin again with the same seeds and nonce; results should be deterministic (same symbols and payout)
	secondSymbols, secondWin, err := eng.Spin(cfg, 100)
	if err != nil {
		t.Fatalf("unexpected error on second spin: %v", err)
	}
	if len(secondSymbols) != 3 {
		t.Fatalf("expected 3 symbols on second spin, got %d", len(secondSymbols))
	}
	for i := 0; i < 3; i++ {
		if firstSymbols[i] != secondSymbols[i] {
			t.Fatalf("fairness spin not deterministic at reel %d: %s vs %s", i, firstSymbols[i], secondSymbols[i])
		}
	}
	if firstWin != secondWin {
		t.Fatalf("fairness payout not deterministic: %d vs %d", firstWin, secondWin)
	}

	_ = firstSymbols
	_ = firstWin
	_ = eng
	_ = cfg
	_ = secondWin
	_ = secondSymbols
}
