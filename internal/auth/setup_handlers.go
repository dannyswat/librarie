package auth

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"librarie/internal/db"
)

// SetupStatusHandler handles GET /api/v1/auth/setup.
// Returns {"needs_setup": true} when no admin user exists yet.
func SetupStatusHandler(q *db.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		hasAdmin, err := q.HasAdminUser(ctx)
		if err != nil {
			slog.Error("setup status: db error", "error", err)
			return echo.ErrInternalServerError
		}
		return c.JSON(http.StatusOK, map[string]bool{"needs_setup": !hasAdmin})
	}
}

type registerAdminRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterAdminHandler handles POST /api/v1/auth/setup.
// Creates the first admin user. Returns 409 Conflict if an admin already exists.
func RegisterAdminHandler(q *db.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		hasAdmin, err := q.HasAdminUser(ctx)
		if err != nil {
			slog.Error("register admin: db error", "error", err)
			return echo.ErrInternalServerError
		}
		if hasAdmin {
			return echo.NewHTTPError(http.StatusConflict, "admin already registered")
		}

		var req registerAdminRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}
		if req.Username == "" || req.Email == "" || req.Password == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "username, email, and password are required")
		}

		hash, err := HashPassword(req.Password)
		if err != nil {
			slog.Error("register admin: hash failed", "error", err)
			return echo.ErrInternalServerError
		}

		user, err := q.CreateUser(ctx, req.Username, req.Email, hash, "admin")
		if err != nil {
			slog.Error("register admin: create user failed", "error", err)
			return echo.ErrInternalServerError
		}

		slog.Info("first admin registered", "username", user.Username)
		return c.JSON(http.StatusCreated, toUserResponse(user))
	}
}
