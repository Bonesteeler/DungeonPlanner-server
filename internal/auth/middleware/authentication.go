package middleware

import (
	"DungeonPlannerServer/internal/auth"
	"strings"

	"github.com/labstack/echo/v4"
)

const authorizationKey = "Authorization"
const prefix = "Bearer "

func CheckAccessToken(tokenManager *auth.TokenManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accessToken := c.Request().Header.Get(authorizationKey)
			if accessToken == "" || !strings.HasPrefix(accessToken, prefix) {
				return echo.ErrUnauthorized
			}
			accessToken = strings.TrimPrefix(accessToken, prefix)
			claims, err := tokenManager.ValidateAccessToken(accessToken)
			if err != nil {
				return echo.ErrUnauthorized
			}
			c.Set(claimsKey, claims)
			return next(c)
		}
	}
}
