package ratelimit

import (
	"sync"
	"time"
)

type TokenBucket struct {
	mu         sync.Mutex
	rate       float64
	capacity   float64
	tokens     float64
	lastRefill time.Time
	lastSeen   time.Time
}

func NewTokenBucket(rps int, burst int) *TokenBucket {
	if rps < 0 {
		rps = 0
	}
	if burst <= 0 {
		burst = rps
	}
	if burst <= 0 {
		burst = 1
	}
	now := time.Now()
	return &TokenBucket{
		rate:       float64(rps),
		capacity:   float64(burst),
		tokens:     float64(burst),
		lastRefill: now,
		lastSeen:   now,
	}
}

func (b *TokenBucket) Allow(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill(now)
	b.lastSeen = now
	if b.rate <= 0 {
		return false
	}
	if b.tokens < 1 {
		return false
	}
	b.tokens -= 1
	return true
}

func (b *TokenBucket) LastSeen() time.Time {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.lastSeen
}

func (b *TokenBucket) refill(now time.Time) {
	if b.lastRefill.IsZero() {
		b.lastRefill = now
		return
	}
	if now.Before(b.lastRefill) {
		b.lastRefill = now
		return
	}
	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed <= 0 {
		return
	}
	b.tokens += elapsed * b.rate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastRefill = now
}
