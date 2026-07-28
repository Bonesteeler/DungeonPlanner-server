package service

import (
	"testing"

	"DungeonPlannerServer/internal/auth"
	"DungeonPlannerServer/internal/auth/repository"
) 

func CreateTestAuthService() *AuthService {
	testRepo := repo.NewInMemoryTokenStore()
	    tokenManager := auth.NewTokenManager(
        []byte("test-access-secret"),
        []byte("test-refresh-secret"),
    )
	return NewAuthService(testRepo, tokenManager)
}

func TestGenerateTokensFromRefreshToken_Success(t *testing.T) {
	testService := CreateTestAuthService()
	testUserID := "user123"
	testRefreshToken, _ := testService.tokenManager.GenerateRefreshToken(testUserID, "user")
	testService.repo.StoreToken(testRefreshToken, testUserID)
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
	testService := CreateTestAuthService()
	invalidToken := "invalid.token.string"
	_, err := testService.GenerateTokensFromRefreshToken(invalidToken)
	if err == nil {
		t.Errorf("expected error for invalid token, got nil")
	}
}

func TestGenerateTokensFromRefreshToken_TokenNotPresent(t *testing.T) {
	testService := CreateTestAuthService()
	testUserID := "user123"
	testRefreshToken, _ := testService.tokenManager.GenerateRefreshToken(testUserID, "user")
	_, err := testService.GenerateTokensFromRefreshToken(testRefreshToken)
	if err == nil {
		t.Errorf("expected error for token not present, got nil")
	}
}
