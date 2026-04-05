package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// IPLimiter tracks per-IP rate limits using token buckets.
type IPLimiter struct {
	mu      sync.Mutex
	buckets map[string]*TokenBucket
	rps     int
	burst   int
	stopCh  chan struct{}
}

// NewIPLimiter creates a per-IP rate limiter.
// rps is the sustained rate; burst is the maximum burst size.
func NewIPLimiter(rps, burst int) *IPLimiter {
	l := &IPLimiter{
		buckets: make(map[string]*TokenBucket),
		rps:     rps,
		burst:   burst,
		stopCh:  make(chan struct{}),
	}
	go func() {
		for {
			select {
			case <-time.After(5 * time.Minute):
				l.cleanup(10 * time.Minute)
			case <-l.stopCh:
				return
			}
		}
	}()
	return l
}

// Stop stops the background cleanup goroutine.
func (l *IPLimiter) Stop() {
	close(l.stopCh)
}

func (l *IPLimiter) Allow(ip string) bool {
	l.mu.Lock()
	b, ok := l.buckets[ip]
	if !ok {
		b = NewTokenBucket(l.rps, l.burst)
		l.buckets[ip] = b
	}
	l.mu.Unlock()
	return b.Allow(time.Now())
}

func (l *IPLimiter) cleanup(maxAge time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	for ip, b := range l.buckets {
		if now.Sub(b.LastSeen()) > maxAge {
			delete(l.buckets, ip)
		}
	}
}

// Middleware returns an http.HandlerFunc wrapper that rate-limits by client IP.
func (l *IPLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !l.Allow(ip) {
			http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

func clientIP(r *http.Request) string {
	// Trust X-Real-IP only if it parses as a valid IP (prevents spoofing).
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		if parsed := net.ParseIP(ip); parsed != nil {
			return ip
		}
	}
	// Fall back to RemoteAddr (strip port).
	host := r.RemoteAddr
	if idx := len(host) - 1; idx > 0 {
		for i := idx; i >= 0; i-- {
			if host[i] == ':' {
				return host[:i]
			}
		}
	}
	return host
}
