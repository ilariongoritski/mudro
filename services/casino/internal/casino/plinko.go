package casino

import (
	"fmt"
	"math"
	"strings"
)

type PlinkoEngine struct {
	draw randFunc
}

func NewPlinkoEngine() *PlinkoEngine {
	return &PlinkoEngine{draw: cryptoDraw}
}

func NewPlinkoEngineWithDraw(draw randFunc) *PlinkoEngine {
	return &PlinkoEngine{draw: draw}
}

func (e *PlinkoEngine) Config() PlinkoConfig {
	return PlinkoConfig{
		Rows:   12,
		Slots:  13,
		MinBet: 1,
		MaxBet: MaxBet(),
		Multipliers: map[PlinkoRisk][]float64{
			PlinkoRiskLow: {
				3.5, 2.0, 1.4, 1.2, 1.1, 1.0, 0.9, 1.0, 1.1, 1.2, 1.4, 2.0, 3.5,
			},
			PlinkoRiskMedium: {
				8.0, 4.0, 2.0, 1.5, 1.2, 1.0, 0.7, 1.0, 1.2, 1.5, 2.0, 4.0, 8.0,
			},
			PlinkoRiskHigh: {
				20.0, 9.0, 4.0, 2.0, 0.5, 0.2, 0.0, 0.2, 0.5, 2.0, 4.0, 9.0, 20.0,
			},
		},
	}
}

func (e *PlinkoEngine) Drop(bet int64, risk PlinkoRisk) (*PlinkoDropResult, error) {
	cfg := e.Config()
	normalizedRisk, multipliers, err := e.resolveRisk(cfg, risk)
	if err != nil {
		return nil, err
	}

	path := make([]int, 0, cfg.Rows)
	slotIndex := 0
	for i := 0; i < cfg.Rows; i++ {
		step := DrawInt(2)
		path = append(path, step)
		slotIndex += step
	}

	multiplier := multipliers[slotIndex]
	payout := int64(math.Round(float64(bet) * multiplier))
	status := "LOST"
	if payout > bet {
		status = "WIN"
	} else if payout > 0 {
		status = "CASHOUT"
	}

	return &PlinkoDropResult{
		Bet:        bet,
		Risk:       normalizedRisk,
		Path:       path,
		Rows:       cfg.Rows,
		SlotIndex:  slotIndex,
		Multiplier: multiplier,
		Payout:     payout,
		NetResult:  payout - bet,
		Status:     status,
		CreatedAt:  nowUTC(),
	}, nil
}

func (e *PlinkoEngine) resolveRisk(cfg PlinkoConfig, risk PlinkoRisk) (PlinkoRisk, []float64, error) {
	switch PlinkoRisk(strings.ToLower(strings.TrimSpace(string(risk)))) {
	case "", PlinkoRiskMedium:
		return PlinkoRiskMedium, cfg.Multipliers[PlinkoRiskMedium], nil
	case PlinkoRiskLow:
		return PlinkoRiskLow, cfg.Multipliers[PlinkoRiskLow], nil
	case PlinkoRiskHigh:
		return PlinkoRiskHigh, cfg.Multipliers[PlinkoRiskHigh], nil
	default:
		return "", nil, fmt.Errorf("unsupported plinko risk %q", risk)
	}
}
