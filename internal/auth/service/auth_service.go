package service

import (
	"errors"

	"github.com/google/uuid"

	"dungeonplannerserver/internal/auth"
	"dungeonplannerserver/internal/auth/model"
)

type TokenRepo interface {
	StoreToken(token string, userID string)
	RemoveToken(token string)
	IsTokenPresent(token string, userID string) bool
}

type UserRepo interface {
	GetPasswordHashByUserId(id uuid.UUID) (string, error)
	StorePasswordHash(id uuid.UUID, passwordHash string) error
	IsEmailExists(email string) (bool, error)
	GetUserIdByEmail(email string) (uuid.UUID, error)
	GetUserRoleByUserId(id uuid.UUID) (string, error)
	AddUser(id uuid.UUID, username string, email string, passwordHash string) error
	GetUserProfileById(id uuid.UUID) (string, string, error)
}

type AuthService struct {
	refreshTokens TokenRepo
	tokenManager  *auth.TokenManager
	users         UserRepo
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
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return model.TokenPair{}, err
	}
	role, err := s.users.GetUserRoleByUserId(userID)
	if err != nil {
		return model.TokenPair{}, err
	}
	newTokenPair, err := s._issueTokenPair(claims.UserID, role)
	if err != nil {
		return model.TokenPair{}, err
	}
	s.refreshTokens.RemoveToken(refreshToken)
	s.refreshTokens.StoreToken(newTokenPair.RefreshToken, claims.UserID)
	return newTokenPair, nil
}

func (s *AuthService) Logout(refreshToken string) error {
	claims, err := s.tokenManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return err
	}
	if !s.refreshTokens.IsTokenPresent(refreshToken, claims.UserID) {
		return errors.New("Unauthorized")
	}
	s.refreshTokens.RemoveToken(refreshToken)
	return nil
}

func (s *AuthService) GenerateTokensFromLogin(email, password string) (model.TokenPair, error) {
	if err := validatePassword(password); err != nil {
		return model.TokenPair{}, err
	}
	userID, err := s.users.GetUserIdByEmail(email)
	if err != nil {
		return model.TokenPair{}, err
	}
	expectedHash, err := s.users.GetPasswordHashByUserId(userID)
	if err != nil {
		return model.TokenPair{}, err
	}
	if !auth.ValidateString(password, expectedHash) {
		return model.TokenPair{}, errors.New("Unauthorized")
	}
	role, err := s.users.GetUserRoleByUserId(userID)
	if err != nil {
		return model.TokenPair{}, err
	}
	tokens, err := s._issueTokenPair(userID.String(), role)
	if err != nil {
		return model.TokenPair{}, err
	}
	s.refreshTokens.StoreToken(tokens.RefreshToken, userID.String())
	return tokens, nil
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

func (s *AuthService) GetUserProfile(userID string) (model.Profile, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return model.Profile{}, err
	}
	username, email, err := s.users.GetUserProfileById(id)
	if err != nil {
		return model.Profile{}, err
	}
	return model.Profile{Username: username, Email: email}, nil
}

func (s *AuthService) UserSignup(username string, password string, email string) (model.TokenPair, error) {
	if err := validatePassword(password); err != nil {
		return model.TokenPair{}, err
	}
	id := uuid.New()
	used, err := s.users.IsEmailExists(email)
	if err != nil {
		return model.TokenPair{}, err
	}
	if used {
		return model.TokenPair{}, errors.New("email already exists in passwords")
	}
	passwordHash := auth.GenerateString(password)
	if err := s.users.AddUser(id, username, email, passwordHash); err != nil {
		return model.TokenPair{}, err
	}
	role, err := s.users.GetUserRoleByUserId(id)
	if err != nil {
		return model.TokenPair{}, err
	}
	tokens, err := s._issueTokenPair(id.String(), role)
	if err != nil {
		return model.TokenPair{}, err
	}
	s.refreshTokens.StoreToken(tokens.RefreshToken, id.String())
	return tokens, nil
}
