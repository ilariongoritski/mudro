package casino

import (
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
)

// fairnessRoll deterministically maps a seed/index pair into [0, max).
// It enables Provably Fair spins when seeds are provided.
func fairnessRoll(serverSeed, clientSeed string, nonce int64, index int, max int) (int, error) {
	if max <= 0 {
		return 0, ErrInvalidConfig
	}
	input := fmt.Sprintf("%s|%s|%d|%d", serverSeed, clientSeed, nonce, index)
	sum := sha256.Sum256([]byte(input))
	v := binary.BigEndian.Uint64(sum[0:8])
	return int(v % uint64(max)), nil
}

// Fairness configuration details are now defined in engine.go

// cryptoDrawFallback provides a tiny crypto-based fallback RNG (no external deps)
func cryptoDrawFallback(max int) (int, error) {
	if max <= 0 {
		return 0, ErrInvalidConfig
	}
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

// Global fairness bridge for non-game-core RNG usage (Roulette, Plinko, Blackjack)
type GlobalFairnessSpec struct {
	ServerSeed  string
	ClientSeed  string
	Nonce       int64
	DrawCounter int64
}

var globalFairnessSpec *GlobalFairnessSpec

func SetGlobalFairnessSpec(ff *GlobalFairnessSpec) {
	globalFairnessSpec = ff
}

func (g *GlobalFairnessSpec) NextRoll(max int) (int, error) {
	if max <= 0 {
		return 0, ErrInvalidConfig
	}
	v, err := fairnessRoll(g.ServerSeed, g.ClientSeed, g.Nonce, int(g.DrawCounter), max)
	if err != nil {
		return 0, err
	}
	g.DrawCounter++
	return v, nil
}

func DrawIntGlobal(max int) int {
	if globalFairnessSpec != nil {
		v, err := globalFairnessSpec.NextRoll(max)
		if err == nil {
			return v
		}
	}
	// fallback to crypto
	v, _ := cryptoDrawFallback(max)
	return v
}

// Backwards-compatible alias for core code expecting DrawInt(max int) int
func DrawInt(max int) int { return DrawIntGlobal(max) }
