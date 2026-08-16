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

	root.POST("/login", func(c echo.Context) error {
		var req dto.LoginRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(400, struct{ Error string }{Error: "Invalid request"})
		}
		return authHandler.GenerateTokensFromLogin(c, req.Username, req.Password)
	})

	root.POST("/signup", func(c echo.Context) error {
		var req dto.SignupRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(400, struct{ Error string }{Error: "Invalid request"})
		}
		return authHandler.UserSignup(c, req.Username, req.Password, req.Email)
	})
}