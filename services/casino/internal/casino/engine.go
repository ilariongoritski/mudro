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
	draw     randFunc
	fairness *Fairness
}

func NewEngine() *Engine {
	return &Engine{draw: cryptoDraw}
}

func NewEngineWithDraw(draw randFunc) *Engine {
	return &Engine{draw: draw}
}

// EnableFairness activates Provably Fair spins for this engine.
func (e *Engine) EnableFairness(serverSeed, clientSeed string, nonce int64) {
	e.fairness = NewFairness(serverSeed, clientSeed, nonce)
}

// DisableFairness disables Provably Fair spins for this engine.
func (e *Engine) DisableFairness() {
	e.fairness = nil
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

	// Prepare 3 spins: either deterministic rolls (provably fair) or random rolls.
	if e.fairness != nil {
		// Reset draw counter for each spin to ensure determinism across identical seed/nonce inputs
		e.fairness.DrawCounter = 0
	}
	symbols := make([]string, 5)
	var rolls [5]int
	if e.fairness != nil {
		// Generate 3 deterministic rolls based on server/client seeds and nonce
		for i := 0; i < 5; i++ {
			r, err := e.fairness.NextRoll(totalWeight)
			if err != nil {
				return nil, 0, err
			}
			rolls[i] = r
		}
	}
	for i := 0; i < 5; i++ {
		var roll int
		if e.fairness != nil {
			roll = rolls[i]
		} else {
			var err error
			roll, err = e.draw(totalWeight)
			if err != nil {
				return nil, 0, err
			}
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
	// Count symbol frequencies
	counts := make(map[string]int)
	for _, s := range symbols {
		counts[s]++
	}
	maxCount := 0
	var maxSym string
	for sym, cnt := range counts {
		if cnt > maxCount {
			maxCount = cnt
			maxSym = sym
		}
	}
	switch maxCount {
	case 5:
		multiplier = cfg.Paytable[maxSym]
	case 4:
		multiplier = cfg.Paytable[maxSym] * 6 / 10
	case 3:
		multiplier = cfg.Paytable[maxSym] * 3 / 10
	case 2:
		multiplier = 1
	}
	if multiplier == 0 && maxCount >= 2 {
		multiplier = 1
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
