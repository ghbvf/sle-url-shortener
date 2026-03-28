package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	maxTokens  = 10
	refillRate = time.Minute // 10 tokens per minute
)

type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// RateLimiter is an IP-based rate limiter using a token bucket algorithm.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	now     func() time.Time // for testing
}

// NewRateLimiter creates a RateLimiter ready for use.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*bucket),
		now:     time.Now,
	}
}

func (rl *RateLimiter) allow(ip string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	b, ok := rl.buckets[ip]
	if !ok {
		rl.buckets[ip] = &bucket{tokens: maxTokens - 1, lastSeen: now}
		return true, 0
	}

	elapsed := now.Sub(b.lastSeen)
	b.tokens += elapsed.Seconds() * (maxTokens / refillRate.Seconds())
	if b.tokens > maxTokens {
		b.tokens = maxTokens
	}
	b.lastSeen = now

	if b.tokens < 1 {
		retryAfter := time.Duration((1 - b.tokens) / (maxTokens / refillRate.Seconds()) * float64(time.Second))
		return false, retryAfter
	}

	b.tokens--
	return true, 0
}

// Wrap returns an http.Handler that rate-limits by client IP before
// delegating to next.
func (rl *RateLimiter) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		allowed, retryAfter := rl.allow(ip)
		if !allowed {
			secs := int(retryAfter.Seconds()) + 1
			w.Header().Set("Retry-After", strconv.Itoa(secs))
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
