package casino

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

type PaytableTier struct {
	MinRoll    int     `json:"minRoll"`
	MaxRoll    int     `json:"maxRoll"`
	Multiplier float64 `json:"multiplier"`
	Label      string  `json:"label"`
	Symbol     string  `json:"symbol"`
}

type PayoutResult struct {
	Multiplier float64
	Amount     float64
	Label      string
	Symbol     string
}

type RtpProfile struct {
	ID        string
	Name      string
	Rtp       float64
	Paytable  []PaytableTier
	IsDefault bool
}

// Cache
var (
	rtpCacheMu sync.RWMutex
	rtpCache   = map[string]*rtpCacheEntry{}
)

type rtpCacheEntry struct {
	profile   *RtpProfile
	expiresAt time.Time
}

const rtpCacheTTL = 60 * time.Second

func EvaluatePayout(roll int, betAmount float64, tiers []PaytableTier) PayoutResult {
	for _, t := range tiers {
		if roll >= t.MinRoll && roll <= t.MaxRoll {
			amount := math.Round(betAmount*t.Multiplier*1e8) / 1e8
			return PayoutResult{
				Multiplier: t.Multiplier,
				Amount:     amount,
				Label:      t.Label,
				Symbol:     t.Symbol,
			}
		}
	}
	return PayoutResult{Label: "МИМО", Symbol: "💀"}
}

func GetActiveRtpProfile(ctx context.Context, repo CasinoRepository, userID string) (*RtpProfile, error) {
	rtpCacheMu.RLock()
	if entry, ok := rtpCache[userID]; ok && time.Now().Before(entry.expiresAt) {
		rtpCacheMu.RUnlock()
		return entry.profile, nil
	}
	rtpCacheMu.RUnlock()

	profile, err := repo.GetActiveRtpProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	rtpCacheMu.Lock()
	rtpCache[userID] = &rtpCacheEntry{profile: profile, expiresAt: time.Now().Add(rtpCacheTTL)}
	rtpCacheMu.Unlock()

	return profile, nil
}

func ClearRtpCache(userID string) {
	rtpCacheMu.Lock()
	if userID == "" {
		rtpCache = map[string]*rtpCacheEntry{}
	} else {
		delete(rtpCache, userID)
	}
	rtpCacheMu.Unlock()
}

func ValidatePaytable(tiers []PaytableTier, targetRtp float64) error {
	covered := make(map[int]bool)
	var calculatedRtp float64

	for _, t := range tiers {
		if t.MinRoll > t.MaxRoll {
			return fmt.Errorf("tier %q: minRoll > maxRoll", t.Label)
		}
		for r := t.MinRoll; r <= t.MaxRoll; r++ {
			if covered[r] {
				return fmt.Errorf("roll %d covered by multiple tiers", r)
			}
			covered[r] = true
		}
		prob := float64(t.MaxRoll-t.MinRoll+1) / 100.0
		calculatedRtp += prob * t.Multiplier
	}

	for r := 0; r <= 99; r++ {
		if !covered[r] {
			return fmt.Errorf("roll %d not covered", r)
		}
	}

	calculatedPct := math.Round(calculatedRtp * 10000) / 100
	diff := math.Abs(calculatedPct - targetRtp)
	if diff > 1 {
		return fmt.Errorf("calculated RTP %.2f%% differs from target %.2f%% by %.2f%%", calculatedPct, targetRtp, diff)
	}

	return nil
}

func ParsePaytable(raw json.RawMessage) ([]PaytableTier, error) {
	var tiers []PaytableTier
	if err := json.Unmarshal(raw, &tiers); err != nil {
		return nil, err
	}
	if len(tiers) == 0 {
		return nil, errors.New("empty paytable")
	}
	return tiers, nil
}
