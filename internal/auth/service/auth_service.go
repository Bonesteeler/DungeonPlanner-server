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

type UserRepo interface {
	GetPasswordHashByEmail(email string) (string, error)
	StorePasswordHash(id uuid.UUID, passwordHash string, email string) error
	IsEmailExists(email string) (bool, error)
	GetUserIdByEmail(email string) (uuid.UUID, error)
	GetUserRoleByUserId(id uuid.UUID) (string, error)
}

type AuthService struct {
	refreshTokens TokenRepo
	tokenManager *auth.TokenManager
	users UserRepo
}

func NewAuthService(tokenRepo TokenRepo, tokenManager *auth.TokenManager, userRepo UserRepo) *AuthService {
	return &AuthService{refreshTokens: tokenRepo, tokenManager: tokenManager, users: userRepo}
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
	expectedHash, err := s.users.GetPasswordHashByEmail(email)
	if err != nil {
		return model.TokenPair{}, err
	}
	if !auth.ValidateString(password, expectedHash) {
		return model.TokenPair{}, errors.New("Unauthorized")
	}
	userID, err := s.users.GetUserIdByEmail(email)
	if err != nil {
		return model.TokenPair{}, err
	}
	role, err := s.users.GetUserRoleByUserId(userID)
	if err != nil {
		return model.TokenPair{}, err
	}
	return s._issueTokenPair(userID.String(), role)
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

func (s *AuthService) UserSignup(username string, password string, email string) error {
	id := uuid.New()
	used, err := s.users.IsEmailExists(email)
	if err != nil {
		return err
	}
	if used {
		return errors.New("email already exists in passwords")
	}
	passwordHash := auth.GenerateString(password)
	return s.users.StorePasswordHash(id, passwordHash, email)
}