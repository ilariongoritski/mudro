package casino

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
)

// ErrInvalidConfig is defined in engine.go to keep a single source of truth

// FairnessRoll implements Stake-grade provably fair result generation.
// It uses HMAC-SHA512(serverSeed, clientSeed:nonce:index)
func FairnessRoll(serverSeed, clientSeed string, nonce int64, index int) ([]byte, error) {
	mac := hmac.New(sha512.New, []byte(serverSeed))
	message := fmt.Sprintf("%s:%d:%d", clientSeed, nonce, index)
	mac.Write([]byte(message))
	return mac.Sum(nil), nil
}

// DrawIntFromHash maps a 512-bit hash slice to an integer in [0, max)
func DrawIntFromHash(hash []byte, max int) int {
	if max <= 0 {
		return 0
	}
	// Use the first 8 bytes for a 64-bit unsigned int
	v := binary.BigEndian.Uint64(hash[0:8])
	return int(v % uint64(max))
}

// GenerateServerSeed generates a secure random 32-byte (64-character hex) string
func GenerateServerSeed() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashServerSeed returns the SHA256 hash of the server seed (used for pre-reveal verification)
func HashServerSeed(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(sum[:])
}

// Fairness is a stateful bridge for game loops (Roulette, Plinko, etc.)
type Fairness struct {
	ServerSeed  string
	ClientSeed  string
	Nonce       int64
	DrawCounter int
}

func (f *Fairness) NextRoll(max int) (int, error) {
	if f == nil {
		return cryptoDrawFallback(max)
	}
	hash, err := FairnessRoll(f.ServerSeed, f.ClientSeed, f.Nonce, f.DrawCounter)
	if err != nil {
		return 0, err
	}
	f.DrawCounter++
	return DrawIntFromHash(hash, max), nil
}

func cryptoDrawFallback(max int) (int, error) {
	if max <= 0 {
		return 0, ErrInvalidConfig
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

func NewFairness(serverSeed, clientSeed string, nonce int64) *Fairness {
	if serverSeed == "" {
		return nil
	}
	return &Fairness{
		ServerSeed:  serverSeed,
		ClientSeed:  clientSeed,
		Nonce:       nonce,
		DrawCounter: 0,
	}
}

func DrawIntWithFairness(fairness *Fairness, max int) int {
	if fairness != nil {
		v, err := fairness.NextRoll(max)
		if err == nil {
			return v
		}
	}
	v, _ := cryptoDrawFallback(max)
	return v
}

// DrawInt is the main entry point for RNG in game logic without fairness context.
func DrawInt(max int) int { return DrawIntWithFairness(nil, max) }
