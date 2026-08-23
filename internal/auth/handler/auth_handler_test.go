package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"DungeonPlannerServer/internal/auth/handler/dto"
	"DungeonPlannerServer/internal/auth/model"
)

// --- Mock ---

type mockAuthService struct {
    generateFromRefreshFn func(refreshToken string) (model.TokenPair, error)
    generateFromLoginFn   func(username, password string) (model.TokenPair, error)
    storePasswordFn func(username string, password string, email string) error
}

func (m *mockAuthService) GenerateTokensFromRefreshToken(refreshToken string) (model.TokenPair, error) {
    return m.generateFromRefreshFn(refreshToken)
}

func (m *mockAuthService) GenerateTokensFromLogin(username, password string) (model.TokenPair, error) {
    return m.generateFromLoginFn(username, password)
}

func (m *mockAuthService) UserSignup(username string, password string, email string) error {
	return m.storePasswordFn(username, password, email)
}

// --- Helpers ---

func newAuthEchoContext(method, path string) (echo.Context, *httptest.ResponseRecorder) {
    e := echo.New()
    req := httptest.NewRequest(method, path, nil)
    rec := httptest.NewRecorder()
    return e.NewContext(req, rec), rec
}

// --- GenerateTokensFromRefreshToken ---

func TestGenerateTokensFromRefreshToken_Success(t *testing.T) {
    expected := model.TokenPair{AccessToken: "access-abc", RefreshToken: "refresh-xyz"}
    h := NewAuthHandler(&mockAuthService{
        generateFromRefreshFn: func(refreshToken string) (model.TokenPair, error) {
            return expected, nil
        },
    })

    c, rec := newAuthEchoContext(http.MethodPost, "/auth/refresh")
    if err := h.GenerateTokensFromRefreshToken(c, "refresh-xyz"); err != nil {
        t.Fatalf("GenerateTokensFromRefreshToken() returned error: %v", err)
    }

    if rec.Code != http.StatusOK {
        t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
    }
    var got dto.GeneratedTokensResponse
    if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
        t.Fatalf("failed to unmarshal response: %v", err)
    }
    if got.AccessToken != expected.AccessToken {
        t.Errorf("AccessToken = %q, want %q", got.AccessToken, expected.AccessToken)
    }
    if got.RefreshToken != expected.RefreshToken {
        t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, expected.RefreshToken)
    }
}

func TestGenerateTokensFromRefreshToken_ServiceError_ReturnsUnauthorized(t *testing.T) {
    h := NewAuthHandler(&mockAuthService{
        generateFromRefreshFn: func(refreshToken string) (model.TokenPair, error) {
            return model.TokenPair{}, errors.New("invalid token")
        },
    })

    c, rec := newAuthEchoContext(http.MethodPost, "/auth/refresh")
    if err := h.GenerateTokensFromRefreshToken(c, "bad-token"); err != nil {
        t.Fatalf("GenerateTokensFromRefreshToken() returned error: %v", err)
    }

    if rec.Code != http.StatusUnauthorized {
        t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
    }
    var got struct{ Error string }
    if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
        t.Fatalf("failed to unmarshal response: %v", err)
    }
    if got.Error != "Unauthorized" {
        t.Errorf("error message = %q, want %q", got.Error, "Unauthorized")
    }
}

// --- GenerateTokensFromLogin ---

func TestGenerateTokensFromLogin_Success(t *testing.T) {
		expected := model.TokenPair{AccessToken: "access-abc", RefreshToken: "refresh-xyz"}
		h := NewAuthHandler(&mockAuthService{
			generateFromLoginFn: func(username, password string) (model.TokenPair, error) {
				return expected, nil
			},
		})
	c, rec := newAuthEchoContext(http.MethodPost, "/auth/login")
	if err := h.GenerateTokensFromLogin(c, "testuser", "password"); err != nil {
		t.Fatalf("GenerateTokensFromLogin() returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got dto.GeneratedTokensResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if got.AccessToken != expected.AccessToken {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, expected.AccessToken)
	}
	if got.RefreshToken != expected.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, expected.RefreshToken)
	}
}

func TestGenerateTokensFromLogin_ServiceError_ReturnsUnauthorized(t *testing.T) {
	h := NewAuthHandler(&mockAuthService{
		generateFromLoginFn: func(username, password string) (model.TokenPair, error) {
			return model.TokenPair{}, errors.New("invalid credentials")
		},
	})
	c, rec := newAuthEchoContext(http.MethodPost, "/auth/login")
	if err := h.GenerateTokensFromLogin(c, "testuser", "wrongpassword"); err != nil {
		t.Fatalf("GenerateTokensFromLogin() returned error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	var got struct{ Error string }
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if got.Error != "Unauthorized" {
		t.Errorf("error message = %q, want %q", got.Error, "Unauthorized")
	}
}

// --- StorePassword ---

func TestStorePassword_Success(t *testing.T) {
	h := NewAuthHandler(&mockAuthService{
		storePasswordFn: func(username string, password string, email string) error {
			return nil
		},
	})
	c, rec := newAuthEchoContext(http.MethodPost, "/auth/store-password")
	if err := h.UserSignup(c, "testuser", "new_password", "test@email.com"); err != nil {
		t.Fatalf("UserSignup() returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestStorePassword_ServiceError(t *testing.T) {
	h := NewAuthHandler(&mockAuthService{
		storePasswordFn: func(username string, password string, email string) error {
			return errors.New("service error")
		},
	})
	c, rec := newAuthEchoContext(http.MethodPost, "/auth/store-password")
	if err := h.UserSignup(c, "testuser", "new_password", "test@email.com"); err != nil {
		t.Fatalf("UserSignup() returned error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}