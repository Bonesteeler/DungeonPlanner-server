package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"DungeonPlannerServer/internal/auth"
	"DungeonPlannerServer/internal/auth/model"
)

var (
	testAccessSecret  = []byte("test-access-secret")
	testRefreshSecret = []byte("test-refresh-secret")
)

func newEchoContext(headerValue string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if headerValue != "" {
		req.Header.Set(authorizationKey, headerValue)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func nextHandler(called *bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		*called = true
		return c.NoContent(http.StatusOK)
	}
}

func TestCheckAccessToken_MissingHeader_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	tm := auth.NewTokenManager(testAccessSecret, testRefreshSecret)
	c, _ := newEchoContext("")

	called := false
	err := CheckAccessToken(tm)(nextHandler(&called))(c)

	if err != echo.ErrUnauthorized {
		t.Errorf("expected echo.ErrUnauthorized, got %v", err)
	}
	if called {
		t.Error("expected next handler not to be called")
	}
}

func TestCheckAccessToken_MissingBearerPrefix_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	tm := auth.NewTokenManager(testAccessSecret, testRefreshSecret)
	tokenStr, err := tm.GenerateAccessToken("user-1", "admin")
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}
	c, _ := newEchoContext(tokenStr)

	called := false
	err = CheckAccessToken(tm)(nextHandler(&called))(c)

	if err != echo.ErrUnauthorized {
		t.Errorf("expected echo.ErrUnauthorized, got %v", err)
	}
	if called {
		t.Error("expected next handler not to be called")
	}
}

func TestCheckAccessToken_InvalidToken_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	tm := auth.NewTokenManager(testAccessSecret, testRefreshSecret)
	c, _ := newEchoContext(prefix + "not-a-valid-token")

	called := false
	err := CheckAccessToken(tm)(nextHandler(&called))(c)

	if err != echo.ErrUnauthorized {
		t.Errorf("expected echo.ErrUnauthorized, got %v", err)
	}
	if called {
		t.Error("expected next handler not to be called")
	}
}

func TestCheckAccessToken_TokenSignedWithWrongSecret_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	tm := auth.NewTokenManager(testAccessSecret, testRefreshSecret)
	otherTm := auth.NewTokenManager([]byte("other-secret"), []byte("other-refresh-secret"))
	tokenStr, err := otherTm.GenerateAccessToken("user-1", "admin")
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}
	c, _ := newEchoContext(prefix + tokenStr)

	called := false
	err = CheckAccessToken(tm)(nextHandler(&called))(c)

	if err != echo.ErrUnauthorized {
		t.Errorf("expected echo.ErrUnauthorized, got %v", err)
	}
	if called {
		t.Error("expected next handler not to be called")
	}
}

func TestCheckAccessToken_ValidToken_CallsNextAndSetsClaims(t *testing.T) {
	t.Parallel()
	tm := auth.NewTokenManager(testAccessSecret, testRefreshSecret)
	tokenStr, err := tm.GenerateAccessToken("user-1", "admin")
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}
	c, rec := newEchoContext(prefix + tokenStr)

	called := false
	err = CheckAccessToken(tm)(nextHandler(&called))(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Error("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	claims, ok := c.Get(claimsKey).(*model.Claims)
	if !ok {
		t.Fatal("expected claims to be set in context as *model.Claims")
	}
	if claims.UserID != "user-1" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-1")
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %q, want %q", claims.Role, "admin")
	}
}
