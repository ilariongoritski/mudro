package casino

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

func GetActiveRtpProfile(ctx context.Context, pool *pgxpool.Pool, userID string) (*RtpProfile, error) {
	rtpCacheMu.RLock()
	if entry, ok := rtpCache[userID]; ok && time.Now().Before(entry.expiresAt) {
		rtpCacheMu.RUnlock()
		return entry.profile, nil
	}
	rtpCacheMu.RUnlock()

	// Check for user-specific assignment
	var profile *RtpProfile
	row := pool.QueryRow(ctx, `
		SELECT p.id, p.name, p.rtp, p.paytable, p.is_default
		FROM casino_rtp_assignments a
		JOIN casino_rtp_profiles p ON p.id = a.rtp_profile_id
		WHERE a.user_id = $1 AND (a.expires_at IS NULL OR a.expires_at > now())
		ORDER BY a.created_at DESC
		LIMIT 1
	`, userID)

	var id, name string
	var rtp float64
	var paytableJSON []byte
	var isDefault bool

	err := row.Scan(&id, &name, &rtp, &paytableJSON, &isDefault)
	if err != nil {
		// Fallback to default
		row = pool.QueryRow(ctx, `
			SELECT id, name, rtp, paytable, is_default
			FROM casino_rtp_profiles
			WHERE is_default = true
			LIMIT 1
		`)
		err = row.Scan(&id, &name, &rtp, &paytableJSON, &isDefault)
		if err != nil {
			return nil, fmt.Errorf("no default RTP profile: %w", err)
		}
	}

	var tiers []PaytableTier
	if err := json.Unmarshal(paytableJSON, &tiers); err != nil {
		return nil, fmt.Errorf("parse paytable: %w", err)
	}

	profile = &RtpProfile{
		ID:        id,
		Name:      name,
		Rtp:       rtp,
		Paytable:  tiers,
		IsDefault: isDefault,
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
