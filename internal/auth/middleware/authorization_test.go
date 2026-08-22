package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"DungeonPlannerServer/internal/auth/model"
)

func newAuthorizationEchoContext() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestCheckRole_NoClaimsInContext_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	c, _ := newAuthorizationEchoContext()

	called := false
	err := CheckRole(User)(nextHandler(&called))(c)

	if err != echo.ErrUnauthorized {
		t.Errorf("expected echo.ErrUnauthorized, got %v", err)
	}
	if called {
		t.Error("expected next handler not to be called")
	}
}

func TestCheckRole_ClaimsWrongType_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	c, _ := newAuthorizationEchoContext()
	c.Set(claimsKey, "not-claims")

	called := false
	err := CheckRole(User)(nextHandler(&called))(c)

	if err != echo.ErrUnauthorized {
		t.Errorf("expected echo.ErrUnauthorized, got %v", err)
	}
	if called {
		t.Error("expected next handler not to be called")
	}
}

func TestCheckRole_NilClaims_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	c, _ := newAuthorizationEchoContext()
	c.Set(claimsKey, (*model.Claims)(nil))

	called := false
	err := CheckRole(User)(nextHandler(&called))(c)

	if err != echo.ErrUnauthorized {
		t.Errorf("expected echo.ErrUnauthorized, got %v", err)
	}
	if called {
		t.Error("expected next handler not to be called")
	}
}

func TestCheckRole_InsufficientRank_ReturnsForbidden(t *testing.T) {
	t.Parallel()
	c, _ := newAuthorizationEchoContext()
	c.Set(claimsKey, &model.Claims{UserID: "user-1", Role: string(User)})

	called := false
	err := CheckRole(Admin)(nextHandler(&called))(c)

	if err != echo.ErrForbidden {
		t.Errorf("expected echo.ErrForbidden, got %v", err)
	}
	if called {
		t.Error("expected next handler not to be called")
	}
}

func TestCheckRole_UnknownRole_ReturnsForbidden(t *testing.T) {
	t.Parallel()
	c, _ := newAuthorizationEchoContext()
	c.Set(claimsKey, &model.Claims{UserID: "user-1", Role: "not-a-real-role"})

	called := false
	err := CheckRole(User)(nextHandler(&called))(c)

	if err != echo.ErrForbidden {
		t.Errorf("expected echo.ErrForbidden, got %v", err)
	}
	if called {
		t.Error("expected next handler not to be called")
	}
}

func TestCheckRole_ExactRank_CallsNext(t *testing.T) {
	t.Parallel()
	c, rec := newAuthorizationEchoContext()
	c.Set(claimsKey, &model.Claims{UserID: "user-1", Role: string(Moderator)})

	called := false
	err := CheckRole(Moderator)(nextHandler(&called))(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Error("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestCheckRole_HigherRank_CallsNext(t *testing.T) {
	t.Parallel()
	c, rec := newAuthorizationEchoContext()
	c.Set(claimsKey, &model.Claims{UserID: "user-1", Role: string(Admin)})

	called := false
	err := CheckRole(User)(nextHandler(&called))(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Error("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
