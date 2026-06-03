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

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authResponse struct {
	User         userResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

func createSessionToken(ctx context.Context, q *db.Queries, userID pgtype.UUID) (string, error) {
	token, err := GenerateToken()
	if err != nil {
		return "", err
	}

	expiresAt := pgtype.Timestamptz{Time: time.Now().Add(SessionDuration), Valid: true}
	if _, err := q.CreateSession(ctx, userID, HashToken(token), expiresAt); err != nil {
		return "", err
	}

	return token, nil
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

		accessToken, err := createSessionToken(ctx, q, user.ID)
		if err != nil {
			slog.Error("login: access token creation failed", "error", err)
			return echo.ErrInternalServerError
		}
		refreshToken, err := createSessionToken(ctx, q, user.ID)
		if err != nil {
			slog.Error("login: refresh token creation failed", "error", err)
			return echo.ErrInternalServerError
		}

		_ = recordLoginAttempt(ctx, q, req.Username, c.RealIP(), req.Password, true)
		SetSessionCookie(c, accessToken)
		return c.JSON(http.StatusOK, authResponse{
			User:         toUserResponse(user),
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		})
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

// RefreshHandler handles POST /api/v1/auth/refresh.
func RefreshHandler(q *db.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req refreshRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}
		if req.RefreshToken == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "refresh_token is required")
		}

		ctx := c.Request().Context()
		session, err := q.GetSessionByTokenHash(ctx, HashToken(req.RefreshToken))
		if err != nil {
			if err == pgx.ErrNoRows {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid refresh token")
			}
			slog.Error("refresh: session lookup failed", "error", err)
			return echo.ErrInternalServerError
		}

		if time.Now().After(session.ExpiresAt.Time) {
			_ = q.DeleteSession(ctx, session.ID)
			return echo.NewHTTPError(http.StatusUnauthorized, "refresh token expired")
		}

		user, err := q.GetUserByID(ctx, session.UserID)
		if err != nil {
			if err == pgx.ErrNoRows {
				return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
			}
			slog.Error("refresh: user lookup failed", "error", err)
			return echo.ErrInternalServerError
		}

		accessToken, err := createSessionToken(ctx, q, user.ID)
		if err != nil {
			slog.Error("refresh: access token creation failed", "error", err)
			return echo.ErrInternalServerError
		}
		newRefreshToken, err := createSessionToken(ctx, q, user.ID)
		if err != nil {
			slog.Error("refresh: refresh token creation failed", "error", err)
			return echo.ErrInternalServerError
		}

		_ = q.DeleteSession(ctx, session.ID)
		SetSessionCookie(c, accessToken)
		return c.JSON(http.StatusOK, authResponse{
			User:         toUserResponse(user),
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken,
		})
	}
}

func recordLoginAttempt(ctx context.Context, q *db.Queries, username, ip, password string, success bool) error {
	fp := PasswordFingerprint(username, password)
	_, err := q.CreateLoginAttempt(ctx, username, ip, &fp, success)
	return err
}
