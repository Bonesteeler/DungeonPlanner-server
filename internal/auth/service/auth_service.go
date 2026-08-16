package service

import (
	"errors"

	"github.com/google/uuid"

	"DungeonPlannerServer/internal/auth"
	"DungeonPlannerServer/internal/auth/model"
)

type TokenRepo interface {
	StoreToken(token string, userID string)
	RemoveToken(token string)
	IsTokenPresent(token string, userID string) bool
}

type PasswordRepo interface {
	GetPasswordHashByEmail(email string) (string, error)
	StorePasswordHash(id uuid.UUID, passwordHash string, email string) error
	IsEmailExists(email string) (bool, error)
}

type AuthService struct {
	refreshTokens TokenRepo
	tokenManager *auth.TokenManager
	passwords PasswordRepo
}

func NewAuthService(tokenRepo TokenRepo, tokenManager *auth.TokenManager, passwordRepo PasswordRepo) *AuthService {
	return &AuthService{refreshTokens: tokenRepo, tokenManager: tokenManager, passwords: passwordRepo}
}

func (s *AuthService) GenerateTokensFromRefreshToken(refreshToken string) (model.TokenPair, error) {
	claims, err := s.tokenManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return model.TokenPair{}, err
	}
	if !s.refreshTokens.IsTokenPresent(refreshToken, claims.UserID) {
		return model.TokenPair{}, errors.New("Unauthorized")
	}
	newTokenPair, err := s._issueTokenPair(claims.UserID, claims.Role)
	if err != nil {
		return model.TokenPair{}, err
	}
	s.refreshTokens.RemoveToken(refreshToken)
	s.refreshTokens.StoreToken(newTokenPair.RefreshToken, claims.UserID)
	return newTokenPair, nil
}

func (s *AuthService) GenerateTokensFromLogin(email, password string) (model.TokenPair, error) {
	expectedHash, err := s.passwords.GetPasswordHashByEmail(email)
	if err != nil {
		return model.TokenPair{}, err
	}
	if !auth.ValidateString(password, expectedHash) {
		return model.TokenPair{}, errors.New("Unauthorized")
	}
	return s._issueTokenPair(email, "user")
}

func (s *AuthService) _issueTokenPair(userID, role string) (model.TokenPair, error) {
	accessToken, err := s.tokenManager.GenerateAccessToken(userID, role)
	if err != nil {
		return model.TokenPair{}, err
	}
	refreshToken, err := s.tokenManager.GenerateRefreshToken(userID, role)
	if err != nil {
		return model.TokenPair{}, err
	}
	return model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) StorePassword(password string, email string) error {
	id := uuid.New()
	used, err := s.passwords.IsEmailExists(email)
	if err != nil {
		return err
	}
	if used {
		return errors.New("email already exists in passwords")
	}
	passwordHash := auth.GenerateString(password)
	return s.passwords.StorePasswordHash(id, passwordHash, email)
}