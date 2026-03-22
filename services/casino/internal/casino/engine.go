package casino

import (
	"errors"
	"math"
	"math/big"
	"sort"

	cryptorand "crypto/rand"
)

var ErrInvalidConfig = errors.New("invalid casino config")

type randFunc func(max int) (int, error)

type Engine struct {
	draw randFunc
}

func NewEngine() *Engine {
	return &Engine{draw: cryptoDraw}
}

func NewEngineWithDraw(draw randFunc) *Engine {
	return &Engine{draw: draw}
}

func (e *Engine) Spin(cfg Config, bet int64) ([]string, int64, error) {
	if bet <= 0 {
		return nil, 0, errors.New("bet must be positive")
	}
	if len(cfg.SymbolWeights) == 0 || len(cfg.Paytable) == 0 || cfg.RTPPercent <= 0 {
		return nil, 0, ErrInvalidConfig
	}

	keys := make([]string, 0, len(cfg.SymbolWeights))
	totalWeight := 0
	for symbol, weight := range cfg.SymbolWeights {
		if weight <= 0 {
			continue
		}
		keys = append(keys, symbol)
		totalWeight += weight
	}
	sort.Strings(keys)
	if len(keys) == 0 || totalWeight <= 0 {
		return nil, 0, ErrInvalidConfig
	}

	symbols := make([]string, 3)
	for i := range symbols {
		roll, err := e.draw(totalWeight)
		if err != nil {
			return nil, 0, err
		}
		cumulative := 0
		for _, symbol := range keys {
			cumulative += cfg.SymbolWeights[symbol]
			if roll < cumulative {
				symbols[i] = symbol
				break
			}
		}
	}

	multiplier := int64(0)
	if symbols[0] == symbols[1] && symbols[1] == symbols[2] {
		multiplier = cfg.Paytable[symbols[0]]
	}

	win := int64(0)
	if multiplier > 0 {
		scaled := float64(multiplier) * cfg.RTPPercent / 100
		if scaled > 0 {
			scaled = math.Round(scaled*100) / 100
		}
		win = int64(math.Round(float64(bet) * scaled))
		if win < bet {
			win = bet
		}
	}

	return symbols, win, nil
}

func cryptoDraw(max int) (int, error) {
	if max <= 0 {
		return 0, ErrInvalidConfig
	}
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}
