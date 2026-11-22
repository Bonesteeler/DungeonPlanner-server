package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"dungeonplannerserver/internal/auth"
	repo "dungeonplannerserver/internal/auth/repository"
)

type mockUserRepo struct {
	GetPasswordHashByUserIdFunc func(id uuid.UUID) (string, error)
	StorePasswordHashFunc func(id uuid.UUID, passwordHash string) error
	IsEmailExistsFunc func(email string) (bool, error)
	GetUserIdByEmailFunc func(email string) (uuid.UUID, error)
	GetUserRoleByUserIdFunc func(id uuid.UUID) (string, error)
	AddUserFunc func(id uuid.UUID, username string, passwordHash string, email string) error
	GetUserProfileByIdFunc func(id uuid.UUID) (string, string, error)
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

func (m *mockUserRepo) GetUserProfileById(id uuid.UUID) (string, string, error) {
	return m.GetUserProfileByIdFunc(id)
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
	expectedPassword := "Password1!"
	expectedPasswordHash := auth.GenerateString(expectedPassword)
	expectedUserID := uuid.New()
	mockRepo := &mockUserRepo{
		GetPasswordHashByUserIdFunc: func(id uuid.UUID) (string, error) {
			return expectedPasswordHash, nil
		},
		GetUserIdByEmailFunc: func(email string) (uuid.UUID, error) {
			return expectedUserID, nil
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
	if !testService.refreshTokens.IsTokenPresent(newTokens.RefreshToken, expectedUserID.String()) {
		t.Errorf("expected refresh token to be stored after login")
	}
}

func TestGenerateTokensFromLogin_InvalidPassword(t *testing.T) {
	mockRepo := &mockUserRepo{
		GetUserIdByEmailFunc: func(email string) (uuid.UUID, error) {
			return uuid.New(), nil
		},
		GetPasswordHashByUserIdFunc: func(id uuid.UUID) (string, error) {
			return "Password1!", nil
		},
	}
	testService := CreateTestAuthService(mockRepo)
	testEmail := "test@email.com"
	invalidPassword := "WrongPassword1!"
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
		GetUserRoleByUserIdFunc: func(id uuid.UUID) (string, error) {
			return "user", nil
		},
	}
	testService := CreateTestAuthService(mockRepo)
	testEmail := "test@email.com"
	tokens, err := testService.UserSignup("testuser", "Password1!", testEmail)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tokens.RefreshToken == "" || tokens.AccessToken == "" {
		t.Errorf("expected valid token pair, got %v", tokens)
	}
	claims, err := testService.tokenManager.ValidateRefreshToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("expected valid refresh token, got %v", err)
	}
	if !testService.refreshTokens.IsTokenPresent(tokens.RefreshToken, claims.UserID) {
		t.Errorf("expected refresh token to be stored after signup")
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
	_, err := testService.UserSignup("testuser", "Password1!", testEmail)
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
	_, err := testService.UserSignup("username", "Password1!", testEmail)
	if err == nil {
		t.Errorf("expected error for repository failure, got nil")
	}
	if err.Error() != databaseError {
		t.Errorf("expected error message '%v', got %v", databaseError, err.Error())
	}
}
func TestGetUserProfile_Success(t *testing.T) {
	expectedUserID := uuid.New()
	mockRepo := &mockUserRepo{
		GetUserProfileByIdFunc: func(id uuid.UUID) (string, string, error) {
			if id != expectedUserID {
				t.Errorf("expected user id %v, got %v", expectedUserID, id)
			}
			return "testuser", "test@email.com", nil
		},
	}
	testService := CreateTestAuthService(mockRepo)
	profile, err := testService.GetUserProfile(expectedUserID.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if profile.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %v", profile.Username)
	}
	if profile.Email != "test@email.com" {
		t.Errorf("expected email 'test@email.com', got %v", profile.Email)
	}
}

func TestGetUserProfile_InvalidUserID(t *testing.T) {
	testService := CreateTestAuthService(&mockUserRepo{})
	_, err := testService.GetUserProfile("not-a-uuid")
	if err == nil {
		t.Errorf("expected error for invalid user id, got nil")
	}
}

func TestGetUserProfile_RepoError(t *testing.T) {
	const databaseError = "database error"
	mockRepo := &mockUserRepo{
		GetUserProfileByIdFunc: func(id uuid.UUID) (string, string, error) {
			return "", "", errors.New(databaseError)
		},
	}
	testService := CreateTestAuthService(mockRepo)
	_, err := testService.GetUserProfile(uuid.New().String())
	if err == nil {
		t.Errorf("expected error for repository failure, got nil")
	}
	if err.Error() != databaseError {
		t.Errorf("expected error message '%v', got %v", databaseError, err.Error())
	}
}


func TestGenerateTokensFromLogin_InvalidPolicyDoesNotReachRepository(t *testing.T) {
	testService := CreateTestAuthService(&mockUserRepo{
		GetUserIdByEmailFunc: func(string) (uuid.UUID, error) {
			t.Fatal("repository called for policy-invalid login")
			return uuid.Nil, nil
		},
	})

	_, err := testService.GenerateTokensFromLogin("test@email.com", "password")
	if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("error = %v, want ErrInvalidPassword", err)
	}
}

func TestUserSignup_InvalidPolicyDoesNotReachRepository(t *testing.T) {
	testService := CreateTestAuthService(&mockUserRepo{
		IsEmailExistsFunc: func(string) (bool, error) {
			t.Fatal("repository called for policy-invalid signup")
			return false, nil
		},
	})

	_, err := testService.UserSignup("testuser", "password", "test@email.com")
	if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("error = %v, want ErrInvalidPassword", err)
	}
}
