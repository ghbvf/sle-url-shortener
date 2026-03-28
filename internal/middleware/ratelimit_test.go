package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestNormalRequestsPass(t *testing.T) {
	rl := NewRateLimiter()
	h := rl.Wrap(okHandler())

	for i := 0; i < maxTokens; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "1.2.3.4:12345"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: got %d, want 200", i+1, rec.Code)
		}
	}
}

func TestRateLimitTriggers429(t *testing.T) {
	rl := NewRateLimiter()
	h := rl.Wrap(okHandler())

	// Exhaust the 10 allowed requests.
	for i := 0; i < maxTokens; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "1.2.3.4:12345"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}

	// The 11th request must be rejected.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("got %d, want 429", rec.Code)
	}
}

func TestRetryAfterHeaderSet(t *testing.T) {
	rl := NewRateLimiter()
	h := rl.Wrap(okHandler())

	for i := 0; i < maxTokens; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:9999"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:9999"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	ra := rec.Header().Get("Retry-After")
	if ra == "" {
		t.Fatal("Retry-After header missing on 429 response")
	}
}

func TestDifferentIPsSeparateLimits(t *testing.T) {
	rl := NewRateLimiter()
	h := rl.Wrap(okHandler())

	// Exhaust limit for IP-A.
	for i := 0; i < maxTokens; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:1111"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}

	// IP-B must still be allowed.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.2:2222"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("different IP got %d, want 200", rec.Code)
	}
}

func TestTokensRefillOverTime(t *testing.T) {
	fakeNow := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rl := NewRateLimiter()
	rl.now = func() time.Time { return fakeNow }
	h := rl.Wrap(okHandler())

	// Exhaust all tokens.
	for i := 0; i < maxTokens; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "5.5.5.5:80"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}

	// Advance time by 6 seconds (= 1 token refilled at 10 per 60s).
	fakeNow = fakeNow.Add(6 * time.Second)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "5.5.5.5:80"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("after refill got %d, want 200", rec.Code)
	}
}
