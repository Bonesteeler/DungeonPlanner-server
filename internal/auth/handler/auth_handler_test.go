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
}

func (m *mockAuthService) GenerateTokensFromRefreshToken(refreshToken string) (model.TokenPair, error) {
    return m.generateFromRefreshFn(refreshToken)
}

func (m *mockAuthService) GenerateTokensFromLogin(username, password string) (model.TokenPair, error) {
    return m.generateFromLoginFn(username, password)
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
