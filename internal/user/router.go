package user

import (
	"librarie/internal/auth"
	"librarie/internal/config"
	"librarie/internal/db"

	"github.com/labstack/echo/v4"
)

// RegisterAdminRoutes mounts phase 3 admin and user-management endpoints.
func RegisterAdminRoutes(g *echo.Group, q *db.Queries, cfg *config.Config) error {
	cipher, err := newCredentialCipher(cfg.EncryptionKey)
	if err != nil {
		return err
	}

	h := &adminHandlers{
		q:      q,
		cipher: cipher,
	}

	sessionMW := auth.SessionMiddleware(q)
	adminMW := auth.RequireRole("admin")

	admin := g.Group("/admin", sessionMW, adminMW)
	admin.GET("/teachers", h.listTeachers)
	admin.POST("/teachers/invite", h.inviteTeacher)
	admin.PUT("/teachers/:id/subjects", h.replaceTeacherSubjects)

	admin.GET("/ai/providers/:provider_key/capabilities", h.listProviderCapabilities)
	admin.PUT("/ai/providers/:provider_key/capabilities/:capability", h.upsertProviderCapability)

	return nil
}
