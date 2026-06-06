package content

import (
	"librarie/internal/auth"
	"librarie/internal/db"
	"librarie/internal/storage"

	"github.com/labstack/echo/v4"
)

// RegisterRoutes mounts all Phase-4 content routes on the given API group
// and registers the upload-serve route on the root Echo instance.
func RegisterRoutes(e *echo.Echo, g *echo.Group, q *db.Queries, store storage.Storage) {
	h := &contentHandlers{q: q, storage: store}

	sessionMW := auth.SessionMiddleware(q)
	teacherMW := auth.RequireRole("teacher")
	adminMW := auth.RequireRole("admin")

	// Upload serving — auth-gated, key must not contain slashes (flat keys)
	e.GET("/uploads/:key", ServeUploadHandler(store), sessionMW)

	// ── Subjects ────────────────────────────────────────────────────────────
	subjects := g.Group("/subjects", sessionMW)
	subjects.GET("", h.listSubjects)
	subjects.POST("", h.createSubject, adminMW)
	subjects.GET("/:id", h.getSubject)
	subjects.PUT("/:id", h.updateSubject, adminMW)
	subjects.DELETE("/:id", h.deleteSubject, adminMW)

	// ── Topics (per subject) ─────────────────────────────────────────────────
	subjects.GET("/:id/topics", h.listSubjectTopics)
	subjects.POST("/:id/topics", h.addSubjectTopic, teacherMW)
	subjects.GET("/:id/contents", h.listContents)

	// ── Contents ─────────────────────────────────────────────────────────────
	contents := g.Group("/contents", sessionMW, teacherMW)
	contents.POST("", h.createContent)
	contents.GET("/:id", h.getContent)
	contents.PUT("/:id", h.updateContent)
	contents.DELETE("/:id", h.deleteContent)
	contents.PUT("/:id/topics", h.replaceContentTopics)

	// ── Pages ─────────────────────────────────────────────────────────────────
	// List/create pages live under /contents/:id/pages
	g.GET("/contents/:id/pages", h.listPages, sessionMW)
	g.POST("/contents/:id/pages", h.createPage, sessionMW, teacherMW)

	// Individual page operations live under /pages/:id
	pages := g.Group("/pages", sessionMW)
	pages.GET("/:id", h.getPage)
	pages.PATCH("/:id", h.patchPage, teacherMW)
	pages.DELETE("/:id", h.deletePage, teacherMW)
	pages.PUT("/:id/blocks", h.replaceBlocks, teacherMW)
}
