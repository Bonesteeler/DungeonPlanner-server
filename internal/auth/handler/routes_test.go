package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"dungeonplannerserver/internal/auth"
	"dungeonplannerserver/internal/auth/model"
)

func TestSetupRoutes_SharesRateLimitAcrossAuthEndpoints(t *testing.T) {
	e := echo.New()
	e.IPExtractor = echo.ExtractIPDirect()
	serviceCalls := 0
	mockService := &mockAuthService{
		generateFromRefreshFn: func(string) (model.TokenPair, error) {
			serviceCalls++
			return model.TokenPair{}, errors.New("invalid token")
		},
		storePasswordFn: func(string, string, string) (model.TokenPair, error) {
			serviceCalls++
			return model.TokenPair{}, errors.New("signup failed")
		},
	}
	SetupRoutes(e, NewAuthHandler(mockService), auth.NewTokenManager(
		[]byte("test-access-secret"),
		[]byte("test-refresh-secret"),
	))

	requests := []struct {
		path string
		body string
	}{
		{path: "/auth/refresh", body: `{"refresh_token":"token"}`},
		{path: "/auth/signup", body: `{"username":"user","email":"test@email.com","password":"Password1!"}`},
	}
	for attempt := 0; attempt < 10; attempt++ {
		request := requests[attempt%len(requests)]
		req := httptest.NewRequest(http.MethodPost, request.path, bytes.NewBufferString(request.body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.RemoteAddr = testClientRemoteAddr
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code == http.StatusTooManyRequests {
			t.Fatalf("request %d was limited before shared burst was exhausted", attempt+1)
		}
	}
	if serviceCalls != 10 {
		t.Fatalf("service calls = %d, want 10", serviceCalls)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(requests[0].body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.RemoteAddr = testClientRemoteAddr
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("request above shared budget status = %d, want 429", rec.Code)
	}
	if serviceCalls != 10 {
		t.Error("rate-limited request reached auth service")
	}

	req = httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(requests[0].body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.RemoteAddr = "203.0.113.11:1234"
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("independent IP status = %d, want 401", rec.Code)
	}
}

func TestSetupRoutes_ProfileStillRequiresAccessToken(t *testing.T) {
	e := echo.New()
	SetupRoutes(e, NewAuthHandler(&mockAuthService{}), auth.NewTokenManager(
		[]byte("test-access-secret"),
		[]byte("test-refresh-secret"),
	))
	req := httptest.NewRequest(http.MethodGet, "/auth/profile", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
