package casino

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type UserRateLimiter struct {
	mu       sync.Mutex
	requests map[int64]*userBucket
	rate     int
	window   time.Duration
	stopCh   chan struct{}
}

type userBucket struct {
	count       int
	windowStart time.Time
}

func NewUserRateLimiter(ctx context.Context, rate int, window time.Duration) *UserRateLimiter {
	limiter := &UserRateLimiter{
		requests: make(map[int64]*userBucket),
		rate:     rate,
		window:   window,
		stopCh:   make(chan struct{}),
	}
	go limiter.cleanupLoop(ctx)
	return limiter
}

func (l *UserRateLimiter) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(l.window)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-l.stopCh:
			return
		case <-ticker.C:
			l.cleanup()
		}
	}
}

func (l *UserRateLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	staleCutoff := now.Add(-2 * l.window)
	for userID, bucket := range l.requests {
		if bucket.windowStart.Before(staleCutoff) {
			delete(l.requests, userID)
		}
	}
}

func (l *UserRateLimiter) Stop() {
	close(l.stopCh)
}

func (l *UserRateLimiter) Allow(userID int64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	bucket, exists := l.requests[userID]

	if !exists || now.Sub(bucket.windowStart) > l.window {
		l.requests[userID] = &userBucket{count: 1, windowStart: now}
		return true
	}

	if bucket.count >= l.rate {
		return false
	}

	bucket.count++
	return true
}

func (h *Handler) rateLimited(actor ParticipantInput) bool {
	if h.userLimiter == nil {
		return false
	}
	return !h.userLimiter.Allow(actor.UserID)
}

func internalAuthMiddleware() func(http.HandlerFunc) http.Handler {
	secret := InternalSecret()
	return func(next http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if secret == "" {
				http.Error(w, `{"error":"casino internal auth is not configured"}`, http.StatusServiceUnavailable)
				return
			}
			if r.Header.Get("X-Internal-Secret") != secret {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			next(w, r)
		})
	}
}
