package auth

import (
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"librarie/internal/db"
)

const invitationTTL = 7 * 24 * time.Hour

// createInvitationRequest is the JSON body for POST /api/v1/invitations.
type createInvitationRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// CreateInvitationHandler handles POST /api/v1/invitations (admin only).
func CreateInvitationHandler(q *db.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		caller := UserFromContext(c)
		if caller.Role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin role required")
		}

		var req createInvitationRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}
		if req.Email == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "email is required")
		}
		if req.Role != "teacher" && req.Role != "admin" && req.Role != "student" {
			return echo.NewHTTPError(http.StatusBadRequest, "role must be admin, teacher, or student")
		}

		token, err := generateInviteToken()
		if err != nil {
			slog.Error("invitation: token generation failed", "error", err)
			return echo.ErrInternalServerError
		}

		expiresAt := pgtype.Timestamptz{Time: time.Now().Add(invitationTTL), Valid: true}

		inv, err := q.CreateInvitation(c.Request().Context(), db.CreateInvitationParams{
			Email:     req.Email,
			Role:      req.Role,
			TokenHash: HashToken(token),
			InvitedBy: caller.ID,
			ExpiresAt: expiresAt,
		})
		if err != nil {
			slog.Error("invitation: DB insert failed", "error", err)
			return echo.ErrInternalServerError
		}

		// Stage 1: log the invite link to stdout instead of sending email.
		slog.Info("invitation created",
			"id", inv.ID.String(),
			"email", req.Email,
			"role", req.Role,
			"token", token, // raw token for local dev
		)

		return c.JSON(http.StatusCreated, map[string]string{
			"id":         inv.ID.String(),
			"email":      inv.Email,
			"role":       inv.Role,
			"expires_at": inv.ExpiresAt.Time.UTC().Format(time.RFC3339),
		})
	}
}

// acceptInvitationRequest is the JSON body for POST /api/v1/invitations/:token/accept.
type acceptInvitationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AcceptInvitationHandler handles POST /api/v1/invitations/:token/accept.
func AcceptInvitationHandler(q *db.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		rawToken := c.Param("token")
		if rawToken == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "token is required")
		}

		var req acceptInvitationRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}
		if req.Username == "" || req.Password == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "username and password are required")
		}
		if len(req.Password) < 8 {
			return echo.NewHTTPError(http.StatusBadRequest, "password must be at least 8 characters")
		}

		ctx := c.Request().Context()

		inv, err := q.GetInvitationByTokenHash(ctx, HashToken(rawToken))
		if err != nil {
			if err == pgx.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound, "invitation not found")
			}
			slog.Error("accept invitation: lookup failed", "error", err)
			return echo.ErrInternalServerError
		}

		if inv.AcceptedAt.Valid {
			return echo.NewHTTPError(http.StatusConflict, "invitation already accepted")
		}
		if time.Now().After(inv.ExpiresAt.Time) {
			return echo.NewHTTPError(http.StatusGone, "invitation has expired")
		}

		hash, err := HashPassword(req.Password)
		if err != nil {
			slog.Error("accept invitation: hash failed", "error", err)
			return echo.ErrInternalServerError
		}

		user, err := q.CreateUser(ctx, req.Username, inv.Email, hash, inv.Role)
		if err != nil {
			slog.Error("accept invitation: create user failed", "error", err)
			// Surface uniqueness violation as a 409.
			return echo.NewHTTPError(http.StatusConflict, "username or email already taken")
		}

		if _, err := q.AcceptInvitation(ctx, inv.ID); err != nil {
			slog.Error("accept invitation: mark accepted failed", "error", err)
			// Non-fatal — user was created; just log.
		}

		return c.JSON(http.StatusCreated, toUserResponse(user))
	}
}

func generateInviteToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
