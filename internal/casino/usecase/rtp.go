package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/goritskimihail/mudro/internal/casino/domain"
	"github.com/goritskimihail/mudro/internal/casino/repository"
)


// Cache
var (
	rtpCacheMu sync.RWMutex
	rtpCache   = map[string]*rtpCacheEntry{}
)

type rtpCacheEntry struct {
	profile   *domain.RtpProfile
	expiresAt time.Time
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
	Paytable  []domain.PaytableTier
	IsDefault bool
}

const rtpCacheTTL = 60 * time.Second

func GetActiveRtpProfile(ctx context.Context, repo repository.CasinoRepository, userID string) (*domain.RtpProfile, error) {
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

func ValidatePaytable(tiers []domain.PaytableTier, targetRtp float64) error {
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

func ParsePaytable(raw json.RawMessage) ([]domain.PaytableTier, error) {
	var tiers []domain.PaytableTier
	if err := json.Unmarshal(raw, &tiers); err != nil {
		return nil, err
	}
	if len(tiers) == 0 {
		return nil, errors.New("empty paytable")
	}
	return tiers, nil
}
