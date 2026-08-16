package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"DungeonPlannerServer/internal/auth"
	repo "DungeonPlannerServer/internal/auth/repository"
)

type mockPasswordRepo struct {
	GetPasswordHashByEmailFunc func(email string) (string, error)
	StorePasswordHashFunc func(id uuid.UUID, passwordHash string, email string) error
	IsEmailExistsFunc func(email string) (bool, error)
}

func (m *mockPasswordRepo) GetPasswordHashByEmail(email string) (string, error) {
	return m.GetPasswordHashByEmailFunc(email)
}

func (m *mockPasswordRepo) StorePasswordHash(id uuid.UUID, passwordHash string, email string) error {
	return m.StorePasswordHashFunc(id, passwordHash, email)
}

func (m *mockPasswordRepo) IsEmailExists(email string) (bool, error) {
	return m.IsEmailExistsFunc(email)
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
		GetPasswordHashByEmailFunc: func(email string) (string, error) {
			return expectedPasswordHash, nil
		},
	}
	testService := CreateTestAuthService(mockRepo)
	testEmail := "test@email.com"
	newTokens, err := testService.GenerateTokensFromLogin(testEmail, expectedPassword)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if newTokens.RefreshToken == "" || newTokens.AccessToken == "" {
		t.Errorf("expected valid token pair, got %v", newTokens)
	}
}

func TestGenerateTokensFromLogin_InvalidPassword(t *testing.T) {
	mockRepo := &mockPasswordRepo{
		GetPasswordHashByEmailFunc: func(email string) (string, error) {
			return "hashed_password", nil
		},
	}
	testService := CreateTestAuthService(mockRepo)
	testEmail := "test@email.com"
	invalidPassword := "wrongpassword"
	_, err := testService.GenerateTokensFromLogin(testEmail, invalidPassword)
	if err == nil {
		t.Errorf("expected error for invalid password, got nil")
	}
}

func TestStorePasswordHash_Success(t *testing.T) {
	mockRepo := &mockPasswordRepo{
		IsEmailExistsFunc: func(email string) (bool, error) {
			return false, nil
		},
		StorePasswordHashFunc: func(id uuid.UUID, passwordHash string, email string) error {
			return nil
		},
	}
	testService := CreateTestAuthService(mockRepo)
	testEmail := "test@email.com"
	err := testService.UserSignup("testuser", "hashed_password", testEmail)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestStorePasswordHash_EmailAlreadyExists(t *testing.T) {
	mockRepo := &mockPasswordRepo{
		IsEmailExistsFunc: func(email string) (bool, error) {
			return true, nil
		},
	}
	testService := CreateTestAuthService(mockRepo)
	testEmail := "test@email.com"
	err := testService.UserSignup("testuser", "hashed_password", testEmail)
	if err == nil {
		t.Errorf("expected error for email already exists, got nil")
	}
	if err.Error() != "email already exists in passwords" {
		t.Errorf("expected error message 'email already exists in passwords', got %v", err.Error())
	}
}

func TestStorePasswordHash_RepoError(t *testing.T) {
	const databaseError = "database error"
	mockRepo := &mockPasswordRepo{
		IsEmailExistsFunc: func(email string) (bool, error) {
			return false, nil
		},
		StorePasswordHashFunc: func(id uuid.UUID, passwordHash string, email string) error {
			return errors.New(databaseError)
		},
	}
	testService := CreateTestAuthService(mockRepo)
	testEmail := "test@email.com"
	err := testService.UserSignup("username", "hashed_password", testEmail)
	if err == nil {
		t.Errorf("expected error for repository failure, got nil")
	}
	if err.Error() != databaseError {
		t.Errorf("expected error message '%v', got %v", databaseError, err.Error())
	}
}