package casino

import "testing"

func TestEngineSpinAppliesPaytableAndRTP(t *testing.T) {
	draws := []int{0, 0, 0}
	engine := NewEngineWithDraw(func(max int) (int, error) {
		v := draws[0]
		draws = draws[1:]
		return v, nil
	})

	cfg := Config{
		RTPPercent:     100,
		InitialBalance: 1000,
		SymbolWeights: map[string]int{
			"cherry": 10,
		},
		Paytable: map[string]int64{
			"cherry": 3,
		},
	}

	symbols, win, err := engine.Spin(cfg, 100)
	if err != nil {
		t.Fatalf("Spin() error = %v", err)
	}
	if win != 300 {
		t.Fatalf("win = %d, want 300", win)
	}
	if len(symbols) != 3 || symbols[0] != "cherry" || symbols[1] != "cherry" || symbols[2] != "cherry" {
		t.Fatalf("unexpected symbols: %#v", symbols)
	}
}
