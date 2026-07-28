package service

import (
	"errors"

	"DungeonPlannerServer/internal/auth"
	"DungeonPlannerServer/internal/auth/model"
)

type AuthRepo interface {
	StoreToken(token string, userID string)
	RemoveToken(token string)
	IsTokenPresent(token string, userID string) bool
}

type AuthService struct {
	repo AuthRepo
	tokenManager *auth.TokenManager
}

func NewAuthService(authRepo AuthRepo, tokenManager *auth.TokenManager) *AuthService {
	return &AuthService{repo: authRepo, tokenManager: tokenManager}
}

func (s *AuthService) GenerateTokensFromRefreshToken(refreshToken string) (model.TokenPair, error) {
	claims, err := s.tokenManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return model.TokenPair{}, err
	}
	if !s.repo.IsTokenPresent(refreshToken, claims.UserID) {
		return model.TokenPair{}, errors.New("Unauthorized")
	}
	newTokenPair, err := s.issueTokenPair(claims.UserID, claims.Role)
	if err != nil {
		return model.TokenPair{}, err
	}
	s.repo.RemoveToken(refreshToken)
	s.repo.StoreToken(newTokenPair.RefreshToken, claims.UserID)
	return newTokenPair, nil
}

func (s *AuthService) GenerateTokensFromLogin(username, password string) (model.TokenPair, error) {
	return model.TokenPair{}, nil
}

func (s *AuthService) issueTokenPair(userID, role string) (model.TokenPair, error) {
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