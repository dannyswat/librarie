package auth

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo/v4"

	"librarie/internal/config"
	"librarie/internal/db"
)

// RegisterRoutes mounts all auth routes under the provided Echo group.
// The group is expected to be /api/v1.
func RegisterRoutes(g *echo.Group, q *db.Queries, cfg *config.Config) error {
	wa, err := webauthn.New(&webauthn.Config{
		RPDisplayName: cfg.RPDisplayName,
		RPID:          cfg.RPID,
		RPOrigins:     []string{cfg.RPOrigin},
	})
	if err != nil {
		return err
	}

	rl := NewRateLimiter(q)
	pk := NewPasskeyHandlers(wa, q)
	sessionMW := SessionMiddleware(q)

	auth := g.Group("/auth")

	// First-run admin setup — disabled once an admin exists
	auth.GET("/setup", SetupStatusHandler(q))
	auth.POST("/setup", RegisterAdminHandler(q), rl.IPRateLimitMiddleware())

	// Public endpoints (rate-limited)
	auth.POST("/login", LoginHandler(q), rl.LoginRateLimitMiddleware())
	auth.POST("/refresh", RefreshHandler(q), rl.IPRateLimitMiddleware())
	auth.POST("/passkey/authenticate/begin", pk.AuthenticateBegin, rl.IPRateLimitMiddleware())
	auth.POST("/passkey/authenticate/complete", pk.AuthenticateComplete, rl.IPRateLimitMiddleware())

	// Invitation acceptance — public
	g.POST("/invitations/:token/accept", AcceptInvitationHandler(q))

	// Authenticated endpoints
	authd := auth
	authd.POST("/logout", LogoutHandler(q), sessionMW)
	authd.GET("/me", MeHandler(), sessionMW)

	// Passkey registration — requires existing session
	auth.POST("/passkey/register/begin", pk.RegisterBegin, sessionMW)
	auth.POST("/passkey/register/complete", pk.RegisterComplete, sessionMW)

	// Invitation creation — requires session (admin check is inside the handler)
	g.POST("/invitations", CreateInvitationHandler(q), sessionMW)

	return nil
}
