package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"dungeonplannerserver/dto"
	"dungeonplannerserver/internal/auth/model"
)

type AuthService interface {
	GenerateTokensFromRefreshToken(refreshToken string) (model.TokenPair, error)
	GenerateTokensFromLogin(username, password string) (model.TokenPair, error)
	UserSignup(username string, password string, email string) (model.TokenPair, error)
	GetUserProfile(userID string) (model.Profile, error)
}

type AuthHandler struct {
	authService AuthService
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) GenerateTokensFromRefreshToken(context echo.Context, refreshToken string) error {
	tokens, err := h.authService.GenerateTokensFromRefreshToken(refreshToken)
	if err != nil {
		return context.JSON(http.StatusUnauthorized, struct{ Error string }{Error: "Unauthorized"})
	}
	return context.JSON(http.StatusOK, dto.GeneratedTokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (h *AuthHandler) GenerateTokensFromLogin(context echo.Context, username, password string) error {
	tokens, err := h.authService.GenerateTokensFromLogin(username, password)
	if err != nil {
		context.Echo().Logger.Error(err)
		return context.JSON(http.StatusUnauthorized, struct{ Error string }{Error: "Unauthorized"})
	}
	return context.JSON(http.StatusOK, dto.GeneratedTokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (h *AuthHandler) UserSignup(context echo.Context, username string, password string, email string) error {
	tokens, err := h.authService.UserSignup(username, password, email)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, struct{ Error string }{Error: "Internal Server Error"})
	}
	return context.JSON(http.StatusOK, dto.GeneratedTokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (h *AuthHandler) GetUserProfile(context echo.Context) error {
	claims, ok := context.Get("claims").(*model.Claims)
	if !ok || claims == nil {
		return context.JSON(http.StatusNotFound, struct{ Error string }{Error: "No data"})
	}
	profile, err := h.authService.GetUserProfile(claims.UserID)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, struct{ Error string }{Error: "Internal Server Error"})
	}
	return context.JSON(http.StatusOK, dto.ProfileResponse{
		Username: profile.Username,
		Email:    profile.Email,
	})
}