package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"dungeonplannerserver/dto"
	"dungeonplannerserver/internal/auth"
	authRepoPck "dungeonplannerserver/internal/auth/repository"
	authServicePck "dungeonplannerserver/internal/auth/service"
)

type integratedUserRepo struct {
	id           uuid.UUID
	username     string
	email        string
	passwordHash string
}

func (r *integratedUserRepo) GetPasswordHashByUserId(id uuid.UUID) (string, error) {
	if id != r.id {
		return "", sql.ErrNoRows
	}
	return r.passwordHash, nil
}

func (r *integratedUserRepo) StorePasswordHash(id uuid.UUID, passwordHash string) error {
	r.id = id
	r.passwordHash = passwordHash
	return nil
}

func (r *integratedUserRepo) IsEmailExists(email string) (bool, error) {
	return r.email == email, nil
}

func (r *integratedUserRepo) GetUserIdByEmail(email string) (uuid.UUID, error) {
	if r.email != email {
		return uuid.Nil, sql.ErrNoRows
	}
	return r.id, nil
}

func (r *integratedUserRepo) GetUserRoleByUserId(id uuid.UUID) (string, error) {
	if id != r.id {
		return "", sql.ErrNoRows
	}
	return "user", nil
}

func (r *integratedUserRepo) AddUser(id uuid.UUID, username, email, passwordHash string) error {
	r.id = id
	r.username = username
	r.email = email
	r.passwordHash = passwordHash
	return nil
}

func (r *integratedUserRepo) GetUserProfileById(id uuid.UUID) (string, string, error) {
	if id != r.id {
		return "", "", sql.ErrNoRows
	}
	return r.username, r.email, nil
}

func newIntegratedAuthRoutes() *echo.Echo {
	e := echo.New()
	tokenManager := auth.NewTokenManager([]byte("test-access-secret"), []byte("test-refresh-secret"))
	authService := authServicePck.NewAuthService(
		authRepoPck.NewInMemoryTokenStore(),
		tokenManager,
		&integratedUserRepo{},
	)
	SetupRoutes(e, NewAuthHandler(authService), tokenManager)
	return e
}

func performIntegratedJSONRequest(t *testing.T, e *echo.Echo, path, body, accessToken string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if accessToken != "" {
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+accessToken)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestAuthRoutes_IntegratedSignupLoginAndProfile(t *testing.T) {
	e := newIntegratedAuthRoutes()
	signupBody := `{"username":"testuser","email":"test@email.com","password":"Password1!"}`

	rec := performIntegratedJSONRequest(t, e, "/auth/signup", signupBody, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("signup status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var signupTokens dto.GeneratedTokensResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &signupTokens); err != nil {
		t.Fatalf("invalid signup response: %v", err)
	}
	if signupTokens.AccessToken == "" || signupTokens.RefreshToken == "" {
		t.Fatal("signup did not return both tokens")
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/profile", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+signupTokens.AccessToken)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("profile status = %d, body = %s", rec.Code, rec.Body.String())
	}

	loginBody := `{"email":"test@email.com","password":"Password1!"}`
	rec = performIntegratedJSONRequest(t, e, "/auth/login", loginBody, "")
	if rec.Code != http.StatusOK {
		t.Errorf("login status = %d, body = %s", rec.Code, rec.Body.String())
	}

	weakLoginBody := `{"email":"test@email.com","password":"password"}`
	rec = performIntegratedJSONRequest(t, e, "/auth/login", weakLoginBody, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("weak login status = %d, want 401", rec.Code)
	}

	wrongLoginBody := `{"email":"test@email.com","password":"WrongPassword1!"}`
	rec = performIntegratedJSONRequest(t, e, "/auth/login", wrongLoginBody, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("wrong login status = %d, want 401", rec.Code)
	}

	weakSignupBody := `{"username":"other","email":"other@email.com","password":"password"}`
	rec = performIntegratedJSONRequest(t, e, "/auth/signup", weakSignupBody, "")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("weak signup status = %d, want 400", rec.Code)
	}
}
