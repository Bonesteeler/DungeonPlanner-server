package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	testAccessSecret  = []byte("test-access-secret")
	testRefreshSecret = []byte("test-refresh-secret")
	testOtherSecret   = []byte("other-secret")
)

// --- GenerateAccessToken ---

func TestGenerateAccessToken_ReturnsNonEmptyToken(t *testing.T) {
	t.Parallel()
	tm := NewTokenManager(testAccessSecret, testRefreshSecret)
	tokenStr, err := tm.GenerateAccessToken("user-1", "admin")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token string")
	}
}

func TestGenerateAccessToken_ClaimsAreCorrect(t *testing.T) {
	t.Parallel()
	tm := NewTokenManager(testAccessSecret, testRefreshSecret)
	tokenStr, err := tm.GenerateAccessToken("user-1", "admin")
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}
	claims, err := tm.ValidateAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("unexpected error validating token: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Errorf("expected UserID %q, got %q", "user-1", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Errorf("expected Role %q, got %q", "admin", claims.Role)
	}
}

// --- GenerateRefreshToken ---

func TestGenerateRefreshToken_ReturnsNonEmptyToken(t *testing.T) {
	t.Parallel()
	tm := NewTokenManager(testAccessSecret, testRefreshSecret)
	tokenStr, err := tm.GenerateRefreshToken("user-1", "admin")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token string")
	}
}

func TestGenerateRefreshToken_ClaimsAreCorrect(t *testing.T) {
	t.Parallel()
	tm := NewTokenManager(testAccessSecret, testRefreshSecret)
	tokenStr, err := tm.GenerateRefreshToken("user-1", "admin")
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}
	claims, err := tm.ValidateRefreshToken(tokenStr)
	if err != nil {
		t.Fatalf("unexpected error validating token: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Errorf("expected UserID %q, got %q", "user-1", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Errorf("expected Role %q, got %q", "admin", claims.Role)
	}
}

// --- ValidateToken (shared table) ---

type validatorEntry struct {
	name          string
	secret        []byte
	generate      func(userID, role string) (string, error)
	validate      func(tokenStr string) (*Claims, error)
	crossGenerate func(userID, role string) (string, error)
}

func TestValidateToken(t *testing.T) {
	t.Parallel()

	tm := NewTokenManager(testAccessSecret, testRefreshSecret)

	validators := []validatorEntry{
		{
			name:          "access",
			secret:        testAccessSecret,
			generate:      tm.GenerateAccessToken,
			validate:      tm.ValidateAccessToken,
			crossGenerate: tm.GenerateRefreshToken,
		},
		{
			name:          "refresh",
			secret:        testRefreshSecret,
			generate:      tm.GenerateRefreshToken,
			validate:      tm.ValidateRefreshToken,
			crossGenerate: tm.GenerateAccessToken,
		},
	}

	for _, v := range validators {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()

			t.Run("ValidToken_ReturnsClaims", func(t *testing.T) {
				t.Parallel()
				tokenStr, err := v.generate("user-42", "player")
				if err != nil {
					t.Fatalf("unexpected generation error: %v", err)
				}
				claims, err := v.validate(tokenStr)
				if err != nil {
					t.Fatalf("unexpected validation error: %v", err)
				}
				if claims.UserID != "user-42" {
					t.Errorf("expected UserID %q, got %q", "user-42", claims.UserID)
				}
				if claims.Role != "player" {
					t.Errorf("expected Role %q, got %q", "player", claims.Role)
				}
			})

			t.Run("ExpiredToken_ReturnsError", func(t *testing.T) {
				t.Parallel()
				expiredToken := generateToken("user-42", "player", -time.Hour)
				tokenStr, err := expiredToken.SignedString(v.secret)
				if err != nil {
					t.Fatalf("unexpected signing error: %v", err)
				}
				_, err = v.validate(tokenStr)
				if err == nil {
					t.Fatal("expected error for expired token, got nil")
				}
			})

			t.Run("WrongSecret_ReturnsError", func(t *testing.T) {
				t.Parallel()
				token := generateToken("user-42", "player", time.Hour)
				tokenStr, err := token.SignedString(testOtherSecret)
				if err != nil {
					t.Fatalf("unexpected signing error: %v", err)
				}
				_, err = v.validate(tokenStr)
				if err == nil {
					t.Fatal("expected error for wrong secret, got nil")
				}
			})

			t.Run("MalformedToken_ReturnsError", func(t *testing.T) {
				t.Parallel()
				_, err := v.validate("not.a.jwt")
				if err == nil {
					t.Fatal("expected error for malformed token, got nil")
				}
			})

			t.Run("WrongSigningMethod_ReturnsError", func(t *testing.T) {
				t.Parallel()
				privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
				if err != nil {
					t.Fatalf("failed to generate RSA key: %v", err)
				}
				claims := Claims{
					UserID: "user-42",
					Role:   "player",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					},
				}
				rs256Token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenStr, err := rs256Token.SignedString(privateKey)
				if err != nil {
					t.Fatalf("unexpected signing error: %v", err)
				}
				_, err = v.validate(tokenStr)
				if err == nil {
					t.Fatal("expected error for wrong signing method, got nil")
				}
			})

			t.Run("CrossTokenRejected", func(t *testing.T) {
				t.Parallel()
				tokenStr, err := v.crossGenerate("user-42", "player")
				if err != nil {
					t.Fatalf("unexpected generation error: %v", err)
				}
				_, err = v.validate(tokenStr)
				if err == nil {
					t.Fatal("expected error when validating opposing token type, got nil")
				}
			})
		})
	}
}
