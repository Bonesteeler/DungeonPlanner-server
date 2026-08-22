package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"DungeonPlannerServer/internal/auth"
	repo "DungeonPlannerServer/internal/auth/repository"
)

type mockUserRepo struct {
	GetPasswordHashByUserIdFunc func(id uuid.UUID) (string, error)
	StorePasswordHashFunc func(id uuid.UUID, passwordHash string) error
	IsEmailExistsFunc func(email string) (bool, error)
	GetUserIdByEmailFunc func(email string) (uuid.UUID, error)
	GetUserRoleByUserIdFunc func(id uuid.UUID) (string, error)
	AddUserFunc func(id uuid.UUID, username string, passwordHash string, email string) error
}

func (m *mockUserRepo) GetPasswordHashByUserId(id uuid.UUID) (string, error) {
	return m.GetPasswordHashByUserIdFunc(id)
}

func (m *mockUserRepo) StorePasswordHash(id uuid.UUID, passwordHash string) error {
	return m.StorePasswordHashFunc(id, passwordHash)
}

func (m *mockUserRepo) IsEmailExists(email string) (bool, error) {
	return m.IsEmailExistsFunc(email)
}

func (m *mockUserRepo) GetUserIdByEmail(email string) (uuid.UUID, error) {
	return m.GetUserIdByEmailFunc(email)
}

func (m *mockUserRepo) GetUserRoleByUserId(id uuid.UUID) (string, error) {
	return m.GetUserRoleByUserIdFunc(id)
}

func (m *mockUserRepo) AddUser(id uuid.UUID, username string, passwordHash string, email string) error {
	return m.AddUserFunc(id, username, passwordHash, email)
}

func CreateTestAuthService(passwordRepo UserRepo) *AuthService {
	testRepo := repo.NewInMemoryTokenStore()
	tokenManager := auth.NewTokenManager(
		[]byte("test-access-secret"),
		[]byte("test-refresh-secret"),
	)
	return NewAuthService(testRepo, tokenManager, passwordRepo)
}

func TestGenerateTokensFromRefreshToken_Success(t *testing.T) {
	testService := CreateTestAuthService(&mockUserRepo{})
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
	testService := CreateTestAuthService(&mockUserRepo{})
	invalidToken := "invalid.token.string"
	_, err := testService.GenerateTokensFromRefreshToken(invalidToken)
	if err == nil {
		t.Errorf("expected error for invalid token, got nil")
	}
}

func TestGenerateTokensFromRefreshToken_TokenNotPresent(t *testing.T) {
	testService := CreateTestAuthService(&mockUserRepo{})
	testUserID := "user123"
	testRefreshToken, _ := testService.tokenManager.GenerateRefreshToken(testUserID, "user")
	_, err := testService.GenerateTokensFromRefreshToken(testRefreshToken)
	if err == nil {
		t.Errorf("expected error for token not present, got nil")
	}
}

func TestGenerateTokensFromLogin_Success(t *testing.T) {
	expectedEmail := "test@email.com"
	expectedPassword := "password"
	expectedPasswordHash := auth.GenerateString(expectedPassword)
	mockRepo := &mockUserRepo{
		GetPasswordHashByUserIdFunc: func(id uuid.UUID) (string, error) {
			return expectedPasswordHash, nil
		},
		GetUserIdByEmailFunc: func(email string) (uuid.UUID, error) {
			return uuid.New(), nil
		},
		GetUserRoleByUserIdFunc: func(id uuid.UUID) (string, error) {
			return "user", nil
		},
	}
	testService := CreateTestAuthService(mockRepo)
	newTokens, err := testService.GenerateTokensFromLogin(expectedEmail, expectedPassword)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if newTokens.RefreshToken == "" || newTokens.AccessToken == "" {
		t.Errorf("expected valid token pair, got %v", newTokens)
	}
}

func TestGenerateTokensFromLogin_InvalidPassword(t *testing.T) {
	mockRepo := &mockUserRepo{
		GetUserIdByEmailFunc: func(email string) (uuid.UUID, error) {
			return uuid.New(), nil
		},
		GetPasswordHashByUserIdFunc: func(id uuid.UUID) (string, error) {
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
	mockRepo := &mockUserRepo{
		AddUserFunc: func(id uuid.UUID, username string, passwordHash string, email string) error {
			return nil
		},
		IsEmailExistsFunc: func(email string) (bool, error) {
			return false, nil
		},
		StorePasswordHashFunc: func(id uuid.UUID, passwordHash string) error {
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
	mockRepo := &mockUserRepo{
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
	mockRepo := &mockUserRepo{
		AddUserFunc: func(id uuid.UUID, username string, passwordHash string, email string) error {
			return errors.New(databaseError)
		},
		IsEmailExistsFunc: func(email string) (bool, error) {
			return false, nil
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