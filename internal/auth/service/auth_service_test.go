package service

import (
	"testing"

	"github.com/google/uuid"

	"DungeonPlannerServer/internal/auth"
	"DungeonPlannerServer/internal/auth/repository"
)

type mockPasswordRepo struct {
	GetPasswordHashByUsernameFunc func(username string) (string, error)
	StorePasswordHashFunc func(id uuid.UUID, passwordHash string, email string) error
}

func (m *mockPasswordRepo) GetPasswordHashByUsername(username string) (string, error) {
	return m.GetPasswordHashByUsernameFunc(username)
}

func (m *mockPasswordRepo) StorePasswordHash(id uuid.UUID, passwordHash string, email string) error {
	return m.StorePasswordHashFunc(id, passwordHash, email)
}


func CreateTestAuthService(passwordRepo PasswordRepo) *AuthService {
	testRepo := repo.NewInMemoryTokenStore()
	tokenManager := auth.NewTokenManager(
		[]byte("test-access-secret"),
		[]byte("test-refresh-secret"),
	)
	return NewAuthService(testRepo, tokenManager, passwordRepo)
}

func TestGenerateTokensFromRefreshToken_Success(t *testing.T) {
	testService := CreateTestAuthService(&mockPasswordRepo{})
	testUserID := "user123"
	testRefreshToken, _ := testService.tokenManager.GenerateRefreshToken(testUserID, "user")
	testService.refreshTokens.StoreToken(testRefreshToken, testUserID)
	newTokens, err := testService.GenerateTokensFromRefreshToken(testRefreshToken)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if newTokens.RefreshToken == "" || newTokens.AccessToken == "" {
		t.Errorf("expected valid token pair, got %v", newTokens)
	}
	if newTokens.RefreshToken == testRefreshToken {
		t.Errorf("expected new refresh token to be different from old one")
	}
}

func TestGenerateTokensFromRefreshToken_InvalidToken(t *testing.T) {
	testService := CreateTestAuthService(&mockPasswordRepo{})
	invalidToken := "invalid.token.string"
	_, err := testService.GenerateTokensFromRefreshToken(invalidToken)
	if err == nil {
		t.Errorf("expected error for invalid token, got nil")
	}
}

func TestGenerateTokensFromRefreshToken_TokenNotPresent(t *testing.T) {
	testService := CreateTestAuthService(&mockPasswordRepo{})
	testUserID := "user123"
	testRefreshToken, _ := testService.tokenManager.GenerateRefreshToken(testUserID, "user")
	_, err := testService.GenerateTokensFromRefreshToken(testRefreshToken)
	if err == nil {
		t.Errorf("expected error for token not present, got nil")
	}
}

func TestGenerateTokensFromLogin_Success(t *testing.T) {
	expectedPassword := "password"
	expectedPasswordHash := auth.GenerateString(expectedPassword)
	mockRepo := &mockPasswordRepo{
		GetPasswordHashByUsernameFunc: func(username string) (string, error) {
			return expectedPasswordHash, nil
		},
	}
	testService := CreateTestAuthService(mockRepo)
	testUsername := "testuser"
	newTokens, err := testService.GenerateTokensFromLogin(testUsername, expectedPassword)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if newTokens.RefreshToken == "" || newTokens.AccessToken == "" {
		t.Errorf("expected valid token pair, got %v", newTokens)
	}
}

func TestGenerateTokensFromLogin_InvalidPassword(t *testing.T) {
	mockRepo := &mockPasswordRepo{
		GetPasswordHashByUsernameFunc: func(username string) (string, error) {
			return "hashed_password", nil
		},
	}
	testService := CreateTestAuthService(mockRepo)
	testUsername := "testuser"
	invalidPassword := "wrongpassword"
	_, err := testService.GenerateTokensFromLogin(testUsername, invalidPassword)
	if err == nil {
		t.Errorf("expected error for invalid password, got nil")
	}
}
