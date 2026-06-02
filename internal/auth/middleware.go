package auth

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"

	"librarie/internal/db"
)

const (
	userContextKey    = "auth_user"
	sessionContextKey = "auth_session"
)

// SessionMiddleware returns an Echo middleware that authenticates requests via
// the session cookie. On success it attaches the user and session to the context.
// On failure it returns 401 Unauthorized.
func SessionMiddleware(q *db.Queries) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, err := GetSessionToken(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
			}

			tokenHash := HashToken(token)
			ctx := c.Request().Context()

			session, err := q.GetSessionByTokenHash(ctx, tokenHash)
			if err != nil {
				if err == pgx.ErrNoRows {
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired session")
				}
				slog.Error("session lookup error", "error", err)
				return echo.ErrInternalServerError
			}

			if time.Now().After(session.ExpiresAt.Time) {
				return echo.NewHTTPError(http.StatusUnauthorized, "session expired")
			}

			user, err := q.GetUserByID(ctx, session.UserID)
			if err != nil {
				if err == pgx.ErrNoRows {
					return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
				}
				slog.Error("user lookup error", "error", err)
				return echo.ErrInternalServerError
			}

			// Update last_seen_at asynchronously to avoid blocking the request.
			go func() {
				if err := q.UpdateSessionLastSeen(ctx, session.ID); err != nil {
					slog.Warn("failed to update session last_seen_at", "error", err)
				}
			}()

			c.Set(userContextKey, user)
			c.Set(sessionContextKey, session)
			return next(c)
		}
	}
}

// UserFromContext retrieves the authenticated user from the Echo context.
// It panics if called outside of an authenticated route.
func UserFromContext(c echo.Context) db.User {
	return c.Get(userContextKey).(db.User)
}

// SessionFromContext retrieves the session from the Echo context.
func SessionFromContext(c echo.Context) db.Session {
	return c.Get(sessionContextKey).(db.Session)
}
