package handler

import (
	"net/http"
	"testing"

	"dungeonplannerserver/internal/auth/model"
	authServicePck "dungeonplannerserver/internal/auth/service"
)

// TEST-NET-3 (203.0.113.0/24) is reserved by RFC 5737 for documentation and examples, not real clients.
const (
	testClientIP         = "203.0.113.10"
	testClientRemoteAddr = testClientIP + ":1234"
)

func performTestLogin(t *testing.T, h *AuthHandler, email string) int {
	t.Helper()
	c, rec := newAuthEchoContext(http.MethodPost, "/auth/login")
	if err := h.GenerateTokensFromLogin(c, email, "Password1!"); err != nil {
		t.Fatalf("GenerateTokensFromLogin() returned error: %v", err)
	}
	return rec.Code
}

func TestGenerateTokensFromLogin_InvalidPasswordPolicyReturnsBadRequest(t *testing.T) {
	h := NewAuthHandler(&mockAuthService{
		generateFromLoginFn: func(string, string) (model.TokenPair, error) {
			return model.TokenPair{}, authServicePck.ErrInvalidPassword
		},
	})

	if status := performTestLogin(t, h, "test@email.com"); status != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", status, http.StatusUnauthorized)
	}
}

func TestUserSignup_InvalidPasswordPolicyReturnsBadRequest(t *testing.T) {
	h := NewAuthHandler(&mockAuthService{
		storePasswordFn: func(string, string, string) (model.TokenPair, error) {
			return model.TokenPair{}, authServicePck.ErrInvalidPassword
		},
	})
	c, rec := newAuthEchoContext(http.MethodPost, "/auth/signup")

	if err := h.UserSignup(c, "testuser", "password", "test@email.com"); err != nil {
		t.Fatalf("UserSignup() returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
