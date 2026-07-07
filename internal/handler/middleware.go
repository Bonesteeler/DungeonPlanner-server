package handler

import (
	"DungeonPlannerServer/internal/auth"

	"github.com/labstack/echo/v4"
)

func CheckAccessToken(tokenManager *auth.TokenManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}
