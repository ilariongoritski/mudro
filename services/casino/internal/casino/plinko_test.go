package casino

import (
	"math"
	"testing"
)

func TestPlinkoDropUsesRiskTable(t *testing.T) {
	engine := NewPlinkoEngine()
	result, err := engine.Drop(100, PlinkoRiskHigh, NewFairness("server-seed", "client-seed", 1))
	if err != nil {
		t.Fatalf("Drop() error = %v", err)
	}

	cfg := engine.Config()
	if result.Risk != PlinkoRiskHigh {
		t.Fatalf("risk = %q, want %q", result.Risk, PlinkoRiskHigh)
	}
	if len(result.Path) != cfg.Rows {
		t.Fatalf("path len = %d, want %d", len(result.Path), cfg.Rows)
	}
	if result.SlotIndex < 0 || result.SlotIndex >= cfg.Slots {
		t.Fatalf("slot index = %d, want 0..%d", result.SlotIndex, cfg.Slots-1)
	}

	wantMultiplier := cfg.Multipliers[PlinkoRiskHigh][result.SlotIndex]
	if result.Multiplier != wantMultiplier {
		t.Fatalf("multiplier = %v, want %v", result.Multiplier, wantMultiplier)
	}

	wantPayout := int64(math.Round(float64(result.Bet) * result.Multiplier))
	if result.Payout != wantPayout {
		t.Fatalf("payout = %d, want %d", result.Payout, wantPayout)
	}
	if result.NetResult != result.Payout-result.Bet {
		t.Fatalf("net result = %d, want %d", result.NetResult, result.Payout-result.Bet)
	}
}

func TestPlinkoDropDefaultsBlankRiskToMedium(t *testing.T) {
	engine := NewPlinkoEngine()
	result, err := engine.Drop(50, "", NewFairness("server-seed", "client-seed", 2))
	if err != nil {
		t.Fatalf("Drop() error = %v", err)
	}
	if result.Risk != PlinkoRiskMedium {
		t.Fatalf("risk = %q, want %q", result.Risk, PlinkoRiskMedium)
	}
}

func TestPlinkoDropRejectsUnsupportedRisk(t *testing.T) {
	engine := NewPlinkoEngine()
	_, err := engine.Drop(25, PlinkoRisk("extreme"), nil)
	if err == nil {
		t.Fatal("expected unsupported risk error")
	}
}
