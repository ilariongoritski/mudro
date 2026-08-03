package casino

import "testing"

func TestSweetBonanzaIsDeterministicForSeedAndNonce(t *testing.T) {
	first := newSweetEngine("server", "client", 7).Spin(10, false)
	second := newSweetEngine("server", "client", 7).Spin(10, false)
	if first.TotalWin != second.TotalWin || len(first.Steps) != len(second.Steps) {
		t.Fatalf("same fairness inputs must produce the same outcome: %#v != %#v", first, second)
	}
	for step, firstStep := range first.Steps {
		if step >= len(second.Steps) || firstStep.Win != second.Steps[step].Win {
			t.Fatalf("cascade %d differs", step)
		}
	}
}

func TestSweetBonanzaBoardShapeAndPayout(t *testing.T) {
	result := newSweetEngine("server", "client", 8).Spin(10, false)
	if len(result.InitialBoard) != sweetReels || len(result.FinalBoard) != sweetReels {
		t.Fatalf("unexpected reel count")
	}
	for reel := range result.InitialBoard {
		if len(result.InitialBoard[reel]) != sweetRows || len(result.FinalBoard[reel]) != sweetRows {
			t.Fatalf("reel %d has invalid row count", reel)
		}
	}
	var summed int64
	for _, step := range result.Steps {
		summed += step.Win
	}
	if result.BombMultiplier == 0 && result.TotalWin != summed {
		t.Fatalf("base-game payout=%d, want summed cascades=%d", result.TotalWin, summed)
	}
}

func TestSweetBonanzaPayoutIsCapped(t *testing.T) {
	const bet int64 = 100
	for nonce := int64(0); nonce < 10_000; nonce++ {
		result := newSweetEngine("server", "client", nonce).Spin(bet, nonce%3 == 0)
		if result.TotalWin < 0 || result.TotalWin > sweetMaxPayoutMultiplier*bet {
			t.Fatalf("nonce %d returned unsafe payout %d", nonce, result.TotalWin)
		}
	}
}

func TestSweetBonanzaFreeSpinBombsOnlyIncreasePayout(t *testing.T) {
	base := newSweetEngine("server", "client", 9).Spin(10, false)
	free := newSweetEngine("server", "client", 9).Spin(10, true)
	if base.TotalWin < 0 || free.TotalWin < 0 {
		t.Fatal("payout must never be negative")
	}
	if free.BombMultiplier > 0 && free.TotalWin == 0 {
		t.Fatal("bomb multiplier must not create a zero payout")
	}
}
