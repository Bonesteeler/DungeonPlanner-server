package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// 203.0.113.x/24 is reserved by RFC 5737 for documentation and examples, not real clients.
const (
	testClientIP         = "203.0.113.10"
	testClientRemoteAddr = testClientIP + ":1234"
	testClientRemoteAddr2 = "203.0.113.11:1234"
)

type stubRateLimiterStore struct {
	allow func(identifier string) (bool, error)
}

func (s stubRateLimiterStore) Allow(identifier string) (bool, error) {
	return s.allow(identifier)
}

func serveRateLimited(t *testing.T, limiter echo.MiddlewareFunc, remoteAddr string) (*httptest.ResponseRecorder, bool) {
	t.Helper()
	e := echo.New()
	e.IPExtractor = echo.ExtractIPDirect()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.RemoteAddr = remoteAddr
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	called := false
	if err := limiter(nextHandler(&called))(c); err != nil {
		t.Fatalf("rate limiter returned error: %v", err)
	}
	return rec, called
}

func TestRateLimiter_AllowedRequestUsesClientIP(t *testing.T) {
	var identifier string
	limiter := NewRateLimiterWithConfig(RateLimiterConfig{
		Store: stubRateLimiterStore{allow: func(id string) (bool, error) {
			identifier = id
			return true, nil
		}},
	})

	rec, called := serveRateLimited(t, limiter, testClientRemoteAddr)
	if rec.Code != http.StatusOK || !called {
		t.Fatalf("allowed request status = %d, called = %v", rec.Code, called)
	}
	if identifier != testClientIP {
		t.Errorf("identifier = %q, want client IP", identifier)
	}
}

func TestRateLimiter_DeniedRequestReturnsJSONWithoutCallingNext(t *testing.T) {
	limiter := NewRateLimiterWithConfig(RateLimiterConfig{
		Store: stubRateLimiterStore{allow: func(string) (bool, error) {
			return false, nil
		}},
		RetryAfterSeconds: 9,
	})

	rec, called := serveRateLimited(t, limiter, testClientRemoteAddr)
	if called {
		t.Error("denied request reached next handler")
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
	if got := rec.Header().Get(echo.HeaderRetryAfter); got != "9" {
		t.Errorf("Retry-After = %q, want 9", got)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if body.Error != "Too many requests" {
		t.Errorf("error = %q, want Too many requests", body.Error)
	}
}

func TestRateLimiter_BurstAndIPIsolation(t *testing.T) {
	limiter := NewRateLimiterWithConfig(RateLimiterConfig{
		Store: echomiddleware.NewRateLimiterMemoryStoreWithConfig(
			echomiddleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Every(time.Hour),
				Burst:     2,
				ExpiresIn: time.Minute,
			},
		),
	})

	for i := 0; i < 2; i++ {
		rec, called := serveRateLimited(t, limiter, testClientRemoteAddr)
		if rec.Code != http.StatusOK || !called {
			t.Fatalf("burst request %d status = %d, called = %v", i+1, rec.Code, called)
		}
	}
	rec, called := serveRateLimited(t, limiter, testClientRemoteAddr)
	if rec.Code != http.StatusTooManyRequests || called {
		t.Errorf("request above burst status = %d, called = %v", rec.Code, called)
	}
	rec, called = serveRateLimited(t, limiter, testClientRemoteAddr2)
	if rec.Code != http.StatusOK || !called {
		t.Errorf("independent IP status = %d, called = %v", rec.Code, called)
	}
}

func TestRateLimiter_ReplenishesSustainedRate(t *testing.T) {
	limiter := NewRateLimiterWithConfig(RateLimiterConfig{
		Store: echomiddleware.NewRateLimiterMemoryStoreWithConfig(
			echomiddleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Every(50 * time.Millisecond),
				Burst:     1,
				ExpiresIn: time.Minute,
			},
		),
	})

	serveRateLimited(t, limiter, testClientRemoteAddr)
	rec, _ := serveRateLimited(t, limiter, testClientRemoteAddr)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("immediate second request status = %d, want 429", rec.Code)
	}
	time.Sleep(60 * time.Millisecond)
	rec, called := serveRateLimited(t, limiter, testClientRemoteAddr)
	if rec.Code != http.StatusOK || !called {
		t.Errorf("replenished request status = %d, called = %v", rec.Code, called)
	}
}
