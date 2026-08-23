package middleware

import (
	"DungeonPlannerServer/internal/auth/model"

	"github.com/labstack/echo/v4"
)

type Role string

const (
	Admin Role = "Admin"
	Moderator Role = "Moderator"
	User  Role = "User"
)

var roleRank = map[Role]int{
	Admin: 3,
	Moderator: 2,
	User:  1,
}

func CheckRole(requiredRole Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, success := c.Get(claimsKey).(*model.Claims)
			if !success || claims == nil {
				return echo.ErrUnauthorized
			}
			if !authorizedByRank(claims.Role, requiredRole) {
				return echo.ErrForbidden
			}
			return next(c)
		}
	}
}

func authorizedByRank(givenRole string, requiredRole Role) bool {
	return roleRank[Role(givenRole)] >= roleRank[requiredRole]
}