package auth

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"

	"DungeonPlannerServer/internal/auth/model"
)

const fifteenMinutes = 15 * time.Minute
const sevenDays = 7 * 24 * time.Hour

type TokenManager struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewTokenManager(accessKey, refreshKey []byte) *TokenManager {
	return &TokenManager{
		accessSecret:  accessKey,
		refreshSecret: refreshKey,
	}
}

func NewTokenManagerFromSecrets() (*TokenManager, error) {
	accessBytes, err := os.ReadFile("/run/secrets/auth-access")
	if err != nil {
		return nil, fmt.Errorf("failed to read access secret file: %w", err)
	}
	accessBytes = bytes.TrimSpace(accessBytes)

	refreshBytes, err := os.ReadFile("/run/secrets/auth-refresh")
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh secret file: %w", err)
	}
	refreshBytes = bytes.TrimSpace(refreshBytes)

	return NewTokenManager(accessBytes, refreshBytes), nil
}

func (tm *TokenManager) GenerateAccessToken(userID string, role string) (string, error) {
	token := generateToken(userID, role, fifteenMinutes)
	return token.SignedString(tm.accessSecret)
}

func (tm *TokenManager) ValidateAccessToken(tokenStr string) (*model.Claims, error) {
	return validateToken(tokenStr, tm.accessSecret)
}

func (tm *TokenManager) GenerateRefreshToken(userID string, role string) (string, error) {
	token := generateToken(userID, role, sevenDays)
	return token.SignedString(tm.refreshSecret)
}

func (tm *TokenManager) ValidateRefreshToken(tokenStr string) (*model.Claims, error) {
	return validateToken(tokenStr, tm.refreshSecret)
}

func validateToken(tokenStr string, secret []byte) (*model.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &model.Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*model.Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func generateToken(userID string, role string, expirationTime time.Duration) *jwt.Token {
	claims := model.Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:			   uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
}
