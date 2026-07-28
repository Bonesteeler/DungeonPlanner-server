package handler

import (
	"github.com/labstack/echo/v4"

	"DungeonPlannerServer/internal/auth"
	"DungeonPlannerServer/internal/auth/handler/dto"
)

func SetupRoutes(e *echo.Echo, authHandler *AuthHandler, tokenManager *auth.TokenManager) {
	root := e.Group("/auth")
	root.POST("/refresh", func(c echo.Context) error {
		var req dto.RefreshTokenRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(400, struct{ Error string }{Error: "Invalid request"})
		}
		return authHandler.GenerateTokensFromRefreshToken(c, req.RefreshToken)
	})
}