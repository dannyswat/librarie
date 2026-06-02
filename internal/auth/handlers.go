package auth

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"librarie/internal/db"
)

// userResponse is the safe public representation of a user (no password hash).
type userResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

func toUserResponse(u db.User) userResponse {
	return userResponse{
		ID:        u.ID.String(),
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt.Time.UTC().Format(time.RFC3339),
	}
}

// loginRequest is the expected JSON body for POST /api/v1/auth/login.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginHandler handles POST /api/v1/auth/login.
func LoginHandler(q *db.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req loginRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}
		if req.Username == "" || req.Password == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "username and password are required")
		}

		ctx := c.Request().Context()

		user, err := q.GetUserByUsername(ctx, req.Username)
		if err != nil {
			if err == pgx.ErrNoRows {
				// Record the attempt and return generic error.
				_ = recordLoginAttempt(ctx, q, req.Username, c.RealIP(), req.Password, false)
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
			}
			slog.Error("login: user lookup failed", "error", err)
			return echo.ErrInternalServerError
		}

		if !CheckPassword(user.PasswordHash, req.Password) {
			_ = recordLoginAttempt(ctx, q, req.Username, c.RealIP(), req.Password, false)
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
		}

		// Create session.
		token, err := GenerateToken()
		if err != nil {
			slog.Error("login: token generation failed", "error", err)
			return echo.ErrInternalServerError
		}

		expiresAt := pgtype.Timestamptz{}
		expiresAt.Time = time.Now().Add(SessionDuration)
		expiresAt.Valid = true

		if _, err := q.CreateSession(ctx, user.ID, HashToken(token), expiresAt); err != nil {
			slog.Error("login: session creation failed", "error", err)
			return echo.ErrInternalServerError
		}

		_ = recordLoginAttempt(ctx, q, req.Username, c.RealIP(), req.Password, true)
		SetSessionCookie(c, token)
		return c.JSON(http.StatusOK, toUserResponse(user))
	}
}

// LogoutHandler handles POST /api/v1/auth/logout.
func LogoutHandler(q *db.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := SessionFromContext(c)
		ctx := c.Request().Context()

		if err := q.DeleteSession(ctx, session.ID); err != nil {
			slog.Error("logout: delete session failed", "error", err)
			return echo.ErrInternalServerError
		}

		ClearSessionCookie(c)
		return c.NoContent(http.StatusNoContent)
	}
}

// MeHandler handles GET /api/v1/auth/me — returns the current user.
func MeHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		user := UserFromContext(c)
		return c.JSON(http.StatusOK, toUserResponse(user))
	}
}

func recordLoginAttempt(ctx context.Context, q *db.Queries, username, ip, password string, success bool) error {
	fp := PasswordFingerprint(username, password)
	_, err := q.CreateLoginAttempt(ctx, username, ip, &fp, success)
	return err
}
