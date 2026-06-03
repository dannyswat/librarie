package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

var roleRank = map[string]int{
	"student": 1,
	"teacher": 2,
	"admin":   3,
}

// RequireRole returns a middleware that enforces a minimum user role.
// Admin satisfies teacher checks because it has a higher rank.
func RequireRole(required string) echo.MiddlewareFunc {
	requiredRank, ok := roleRank[required]
	if !ok {
		panic("auth.RequireRole: unknown role " + required)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := UserFromContext(c)
			if roleRank[user.Role] < requiredRank {
				return echo.NewHTTPError(http.StatusForbidden, required+" role required")
			}
			return next(c)
		}
	}
}
