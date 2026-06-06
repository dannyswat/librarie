package content

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"librarie/internal/auth"
	"librarie/internal/db"
	"librarie/internal/storage"
)

type contentHandlers struct {
	q       *db.Queries
	storage storage.Storage
}

// ─── UUID helper ───────────────────────────────────────────────────────────

func parseUUIDParam(value string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := id.Scan(value)
	return id, err
}

// ─── Storage key helper ─────────────────────────────────────────────────────

// newStorageKey generates a random 32-hex-char key with the given extension.
func newStorageKey(ext string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return fmt.Sprintf("%x%s", b, ext), nil
}

// ─── Teacher ownership check ─────────────────────────────────────────────────

func (h *contentHandlers) teacherOwnsSubject(c echo.Context, subjectID pgtype.UUID) (bool, error) {
	user := auth.UserFromContext(c)
	if user.Role == "admin" {
		return true, nil
	}
	subjects, err := h.q.ListSubjectsByTeacher(c.Request().Context(), user.ID)
	if err != nil {
		return false, err
	}
	for _, s := range subjects {
		if s.ID == subjectID {
			return true, nil
		}
	}
	return false, nil
}

// ─── Response types ──────────────────────────────────────────────────────────

type subjectResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	CoverImageKey *string `json:"cover_image_key,omitempty"`
	CoverImageURL *string `json:"cover_image_url,omitempty"`
	Position      int32   `json:"position"`
	CreatedAt     string  `json:"created_at"`
}

type topicResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

type contentResponse struct {
	ID          string          `json:"id"`
	SubjectID   string          `json:"subject_id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Position    int32           `json:"position"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
	Topics      []topicResponse `json:"topics,omitempty"`
}

type pageResponse struct {
	ID        string          `json:"id"`
	ContentID string          `json:"content_id"`
	Name      string          `json:"name"`
	Position  int32           `json:"position"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
	Blocks    []blockResponse `json:"blocks,omitempty"`
}

type blockResponse struct {
	ID        string          `json:"id"`
	PageID    string          `json:"page_id"`
	Type      string          `json:"type"`
	Position  int32           `json:"position"`
	Data      json.RawMessage `json:"data"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

// ─── Converters ──────────────────────────────────────────────────────────────

func (h *contentHandlers) toSubjectResponse(s db.Subject) subjectResponse {
	r := subjectResponse{
		ID:          s.ID.String(),
		Name:        s.Name,
		Description: s.Description,
		Position:    s.Position,
		CreatedAt:   s.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if s.CoverImageKey != nil {
		key := *s.CoverImageKey
		r.CoverImageKey = &key
		url := h.storage.URL(key)
		r.CoverImageURL = &url
	}
	return r
}

func toTopicResponse(t db.Topic) topicResponse {
	return topicResponse{
		ID:          t.ID.String(),
		Name:        t.Name,
		Description: t.Description,
		CreatedAt:   t.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
}

func toContentResponse(c db.Content, topics []db.Topic) contentResponse {
	r := contentResponse{
		ID:          c.ID.String(),
		SubjectID:   c.SubjectID.String(),
		Title:       c.Title,
		Description: c.Description,
		Position:    c.Position,
		CreatedAt:   c.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   c.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if len(topics) > 0 {
		r.Topics = make([]topicResponse, len(topics))
		for i, t := range topics {
			r.Topics[i] = toTopicResponse(t)
		}
	}
	return r
}

func toPageResponse(p db.Page, blocks []db.Block) pageResponse {
	r := pageResponse{
		ID:        p.ID.String(),
		ContentID: p.ContentID.String(),
		Name:      p.Name,
		Position:  p.Position,
		CreatedAt: p.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: p.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if len(blocks) > 0 {
		r.Blocks = make([]blockResponse, len(blocks))
		for i, b := range blocks {
			r.Blocks[i] = toBlockResponse(b)
		}
	}
	return r
}

func toBlockResponse(b db.Block) blockResponse {
	data := json.RawMessage(b.Data)
	if len(data) == 0 {
		data = json.RawMessage("{}")
	}
	return blockResponse{
		ID:        b.ID.String(),
		PageID:    b.PageID.String(),
		Type:      b.Type,
		Position:  b.Position,
		Data:      data,
		CreatedAt: b.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: b.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
}

// ─── Subjects handlers ───────────────────────────────────────────────────────

func (h *contentHandlers) listSubjects(c echo.Context) error {
	ctx := c.Request().Context()
	user := auth.UserFromContext(c)

	var subjects []db.Subject
	var err error
	if user.Role == "admin" {
		subjects, err = h.q.ListSubjects(ctx)
	} else {
		subjects, err = h.q.ListSubjectsByTeacher(ctx, user.ID)
	}
	if err != nil {
		slog.Error("list subjects failed", "error", err)
		return echo.ErrInternalServerError
	}

	resp := make([]subjectResponse, len(subjects))
	for i, s := range subjects {
		resp[i] = h.toSubjectResponse(s)
	}
	return c.JSON(http.StatusOK, map[string]any{"subjects": resp})
}

func (h *contentHandlers) createSubject(c echo.Context) error {
	ctx := c.Request().Context()
	user := auth.UserFromContext(c)

	name := c.FormValue("name")
	description := c.FormValue("description")
	if strings.TrimSpace(name) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}

	// Compute next position
	all, err := h.q.ListSubjects(ctx)
	if err != nil {
		slog.Error("list subjects for position failed", "error", err)
		return echo.ErrInternalServerError
	}
	var maxPos int32
	for _, s := range all {
		if s.Position > maxPos {
			maxPos = s.Position
		}
	}

	var coverKey *string
	if file, header, ferr := c.Request().FormFile("cover_image"); ferr == nil {
		defer file.Close()
		ext := mediaExt(header.Filename, header.Header.Get("Content-Type"))
		key, kerr := newStorageKey(ext)
		if kerr != nil {
			return echo.ErrInternalServerError
		}
		if perr := h.storage.Put(ctx, key, file, header.Size); perr != nil {
			slog.Error("store cover image failed", "error", perr)
			return echo.ErrInternalServerError
		}
		coverKey = &key
	}

	subject, err := h.q.CreateSubject(ctx, db.CreateSubjectParams{
		Name:          strings.TrimSpace(name),
		Description:   strings.TrimSpace(description),
		CoverImageKey: coverKey,
		Position:      maxPos + 1,
		CreatedBy:     user.ID,
	})
	if err != nil {
		slog.Error("create subject failed", "error", err)
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusCreated, map[string]any{"subject": h.toSubjectResponse(subject)})
}

func (h *contentHandlers) getSubject(c echo.Context) error {
	id, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid subject id")
	}

	subject, err := h.q.GetSubjectByID(c.Request().Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "subject not found")
		}
		slog.Error("get subject failed", "error", err)
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusOK, map[string]any{"subject": h.toSubjectResponse(subject)})
}

func (h *contentHandlers) updateSubject(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid subject id")
	}

	existing, err := h.q.GetSubjectByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "subject not found")
		}
		slog.Error("get subject failed", "error", err)
		return echo.ErrInternalServerError
	}

	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		name = existing.Name
	}
	description := c.FormValue("description")
	if description == "" {
		description = existing.Description
	}
	posStr := c.FormValue("position")
	position := existing.Position
	if posStr != "" {
		if _, scanErr := fmt.Sscanf(posStr, "%d", &position); scanErr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid position")
		}
	}

	coverKey := existing.CoverImageKey
	if file, header, ferr := c.Request().FormFile("cover_image"); ferr == nil {
		defer file.Close()
		// Delete old image if present
		if coverKey != nil {
			_ = h.storage.Delete(ctx, *coverKey)
		}
		ext := mediaExt(header.Filename, header.Header.Get("Content-Type"))
		key, kerr := newStorageKey(ext)
		if kerr != nil {
			return echo.ErrInternalServerError
		}
		if perr := h.storage.Put(ctx, key, file, header.Size); perr != nil {
			slog.Error("store cover image failed", "error", perr)
			return echo.ErrInternalServerError
		}
		coverKey = &key
	}

	updated, err := h.q.UpdateSubject(ctx, db.UpdateSubjectParams{
		ID:            id,
		Name:          name,
		Description:   description,
		CoverImageKey: coverKey,
		Position:      position,
	})
	if err != nil {
		slog.Error("update subject failed", "error", err)
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusOK, map[string]any{"subject": h.toSubjectResponse(updated)})
}

func (h *contentHandlers) deleteSubject(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid subject id")
	}

	subject, err := h.q.GetSubjectByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "subject not found")
		}
		slog.Error("get subject failed", "error", err)
		return echo.ErrInternalServerError
	}

	if err := h.q.DeleteSubject(ctx, id); err != nil {
		slog.Error("delete subject failed", "error", err)
		return echo.ErrInternalServerError
	}

	// Best-effort cover image cleanup
	if subject.CoverImageKey != nil {
		_ = h.storage.Delete(ctx, *subject.CoverImageKey)
	}

	return c.NoContent(http.StatusNoContent)
}

// ─── Topics handlers ─────────────────────────────────────────────────────────

func (h *contentHandlers) listSubjectTopics(c echo.Context) error {
	id, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid subject id")
	}

	topics, err := h.q.ListTopicsBySubject(c.Request().Context(), id)
	if err != nil {
		slog.Error("list subject topics failed", "error", err)
		return echo.ErrInternalServerError
	}

	resp := make([]topicResponse, len(topics))
	for i, t := range topics {
		resp[i] = toTopicResponse(t)
	}
	return c.JSON(http.StatusOK, map[string]any{"topics": resp})
}

type addTopicRequest struct {
	// Provide either `id` to link an existing topic, or `name`+`description` to create new.
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *contentHandlers) addSubjectTopic(c echo.Context) error {
	ctx := c.Request().Context()
	subjectID, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid subject id")
	}

	// Verify subject exists
	if _, err := h.q.GetSubjectByID(ctx, subjectID); err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "subject not found")
		}
		slog.Error("get subject failed", "error", err)
		return echo.ErrInternalServerError
	}

	owns, err := h.teacherOwnsSubject(c, subjectID)
	if err != nil {
		slog.Error("check teacher ownership failed", "error", err)
		return echo.ErrInternalServerError
	}
	if !owns {
		return echo.NewHTTPError(http.StatusForbidden, "not assigned to this subject")
	}

	var req addTopicRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	var topic db.Topic

	if req.ID != "" {
		// Link existing topic
		topicID, perr := parseUUIDParam(req.ID)
		if perr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid topic id")
		}
		topic, err = h.q.GetTopicByID(ctx, topicID)
		if err != nil {
			if err == pgx.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound, "topic not found")
			}
			slog.Error("get topic failed", "error", err)
			return echo.ErrInternalServerError
		}
	} else {
		// Create new topic
		if strings.TrimSpace(req.Name) == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "name or id is required")
		}
		user := auth.UserFromContext(c)
		topic, err = h.q.CreateTopic(ctx, strings.TrimSpace(req.Name), strings.TrimSpace(req.Description), user.ID)
		if err != nil {
			slog.Error("create topic failed", "error", err)
			return echo.ErrInternalServerError
		}
	}

	// Compute position for subject_topics
	existing, err := h.q.ListTopicsBySubject(ctx, subjectID)
	if err != nil {
		slog.Error("list topics for position failed", "error", err)
		return echo.ErrInternalServerError
	}
	pos := int32(len(existing) + 1)

	if err := h.q.AddTopicToSubject(ctx, subjectID, topic.ID, pos); err != nil {
		slog.Error("add topic to subject failed", "error", err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusCreated, map[string]any{"topic": toTopicResponse(topic)})
}

// ─── Contents handlers ───────────────────────────────────────────────────────

func (h *contentHandlers) listContents(c echo.Context) error {
	ctx := c.Request().Context()
	subjectID, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid subject id")
	}

	contents, err := h.q.ListContentsBySubject(ctx, subjectID)
	if err != nil {
		slog.Error("list contents failed", "error", err)
		return echo.ErrInternalServerError
	}

	resp := make([]contentResponse, len(contents))
	for i, c2 := range contents {
		resp[i] = toContentResponse(c2, nil)
	}
	return c.JSON(http.StatusOK, map[string]any{"contents": resp})
}

type createContentRequest struct {
	SubjectID   string `json:"subject_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *contentHandlers) createContent(c echo.Context) error {
	ctx := c.Request().Context()
	user := auth.UserFromContext(c)

	var req createContentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if strings.TrimSpace(req.Title) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}
	subjectID, err := parseUUIDParam(req.SubjectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid subject_id")
	}

	owns, err := h.teacherOwnsSubject(c, subjectID)
	if err != nil {
		slog.Error("check teacher ownership failed", "error", err)
		return echo.ErrInternalServerError
	}
	if !owns {
		return echo.NewHTTPError(http.StatusForbidden, "not assigned to this subject")
	}

	existing, err := h.q.ListContentsBySubject(ctx, subjectID)
	if err != nil {
		slog.Error("list contents for position failed", "error", err)
		return echo.ErrInternalServerError
	}
	var maxPos int32
	for _, c2 := range existing {
		if c2.Position > maxPos {
			maxPos = c2.Position
		}
	}

	content, err := h.q.CreateContent(ctx, db.CreateContentParams{
		SubjectID:   subjectID,
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Position:    maxPos + 1,
		CreatedBy:   user.ID,
	})
	if err != nil {
		slog.Error("create content failed", "error", err)
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusCreated, map[string]any{"content": toContentResponse(content, nil)})
}

func (h *contentHandlers) getContent(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid content id")
	}

	content, err := h.q.GetContentByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "content not found")
		}
		slog.Error("get content failed", "error", err)
		return echo.ErrInternalServerError
	}

	topics, err := h.q.ListTopicsByContent(ctx, id)
	if err != nil {
		slog.Error("list content topics failed", "error", err)
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusOK, map[string]any{"content": toContentResponse(content, topics)})
}

type updateContentRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Position    int32  `json:"position"`
}

func (h *contentHandlers) updateContent(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid content id")
	}

	existing, err := h.q.GetContentByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "content not found")
		}
		slog.Error("get content failed", "error", err)
		return echo.ErrInternalServerError
	}

	owns, err := h.teacherOwnsSubject(c, existing.SubjectID)
	if err != nil {
		slog.Error("check teacher ownership failed", "error", err)
		return echo.ErrInternalServerError
	}
	if !owns {
		return echo.NewHTTPError(http.StatusForbidden, "not assigned to this subject")
	}

	var req updateContentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if strings.TrimSpace(req.Title) == "" {
		req.Title = existing.Title
	}
	if req.Description == "" {
		req.Description = existing.Description
	}
	if req.Position == 0 {
		req.Position = existing.Position
	}

	updated, err := h.q.UpdateContent(ctx, id, strings.TrimSpace(req.Title), req.Description, req.Position)
	if err != nil {
		slog.Error("update content failed", "error", err)
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusOK, map[string]any{"content": toContentResponse(updated, nil)})
}

func (h *contentHandlers) deleteContent(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid content id")
	}

	existing, err := h.q.GetContentByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "content not found")
		}
		slog.Error("get content failed", "error", err)
		return echo.ErrInternalServerError
	}

	owns, err := h.teacherOwnsSubject(c, existing.SubjectID)
	if err != nil {
		slog.Error("check teacher ownership failed", "error", err)
		return echo.ErrInternalServerError
	}
	if !owns {
		return echo.NewHTTPError(http.StatusForbidden, "not assigned to this subject")
	}

	if err := h.q.DeleteContent(ctx, id); err != nil {
		slog.Error("delete content failed", "error", err)
		return echo.ErrInternalServerError
	}
	return c.NoContent(http.StatusNoContent)
}

type replaceTopicsRequest struct {
	TopicIDs []string `json:"topic_ids"`
}

func (h *contentHandlers) replaceContentTopics(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid content id")
	}

	existing, err := h.q.GetContentByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "content not found")
		}
		slog.Error("get content failed", "error", err)
		return echo.ErrInternalServerError
	}

	owns, err := h.teacherOwnsSubject(c, existing.SubjectID)
	if err != nil {
		slog.Error("check teacher ownership failed", "error", err)
		return echo.ErrInternalServerError
	}
	if !owns {
		return echo.NewHTTPError(http.StatusForbidden, "not assigned to this subject")
	}

	var req replaceTopicsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Delete existing associations
	if err := h.q.ReplaceContentTopics(ctx, id); err != nil {
		slog.Error("replace content topics delete failed", "error", err)
		return echo.ErrInternalServerError
	}

	// Re-insert
	var addedTopics []db.Topic
	for _, rawID := range req.TopicIDs {
		topicID, perr := parseUUIDParam(rawID)
		if perr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid topic id: "+rawID)
		}
		t, terr := h.q.GetTopicByID(ctx, topicID)
		if terr != nil {
			if terr == pgx.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound, "topic not found: "+rawID)
			}
			slog.Error("get topic failed", "error", terr)
			return echo.ErrInternalServerError
		}
		if err := h.q.AddTopicToContent(ctx, id, topicID); err != nil {
			slog.Error("add topic to content failed", "error", err)
			return echo.ErrInternalServerError
		}
		addedTopics = append(addedTopics, t)
	}

	resp := make([]topicResponse, len(addedTopics))
	for i, t := range addedTopics {
		resp[i] = toTopicResponse(t)
	}
	return c.JSON(http.StatusOK, map[string]any{"topics": resp})
}

// ─── Pages handlers ──────────────────────────────────────────────────────────

func (h *contentHandlers) listPages(c echo.Context) error {
	ctx := c.Request().Context()
	contentID, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid content id")
	}

	pages, err := h.q.ListPagesByContent(ctx, contentID)
	if err != nil {
		slog.Error("list pages failed", "error", err)
		return echo.ErrInternalServerError
	}

	resp := make([]pageResponse, len(pages))
	for i, p := range pages {
		resp[i] = toPageResponse(p, nil)
	}
	return c.JSON(http.StatusOK, map[string]any{"pages": resp})
}

type createPageRequest struct {
	Name string `json:"name"`
}

func (h *contentHandlers) createPage(c echo.Context) error {
	ctx := c.Request().Context()
	user := auth.UserFromContext(c)
	contentID, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid content id")
	}

	content, err := h.q.GetContentByID(ctx, contentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "content not found")
		}
		slog.Error("get content failed", "error", err)
		return echo.ErrInternalServerError
	}

	owns, err := h.teacherOwnsSubject(c, content.SubjectID)
	if err != nil {
		slog.Error("check teacher ownership failed", "error", err)
		return echo.ErrInternalServerError
	}
	if !owns {
		return echo.NewHTTPError(http.StatusForbidden, "not assigned to this subject")
	}

	var req createPageRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if strings.TrimSpace(req.Name) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}

	existing, err := h.q.ListPagesByContent(ctx, contentID)
	if err != nil {
		slog.Error("list pages for position failed", "error", err)
		return echo.ErrInternalServerError
	}
	var maxPos int32
	for _, p := range existing {
		if p.Position > maxPos {
			maxPos = p.Position
		}
	}

	page, err := h.q.CreatePage(ctx, contentID, strings.TrimSpace(req.Name), maxPos+1, user.ID)
	if err != nil {
		slog.Error("create page failed", "error", err)
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusCreated, map[string]any{"page": toPageResponse(page, nil)})
}

func (h *contentHandlers) getPage(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid page id")
	}

	page, err := h.q.GetPageByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "page not found")
		}
		slog.Error("get page failed", "error", err)
		return echo.ErrInternalServerError
	}

	blocks, err := h.q.ListBlocksByPage(ctx, id)
	if err != nil {
		slog.Error("list blocks failed", "error", err)
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusOK, map[string]any{"page": toPageResponse(page, blocks)})
}

type patchPageRequest struct {
	Name     string `json:"name"`
	Position int32  `json:"position"`
}

func (h *contentHandlers) patchPage(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid page id")
	}

	page, err := h.q.GetPageByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "page not found")
		}
		slog.Error("get page failed", "error", err)
		return echo.ErrInternalServerError
	}

	content, err := h.q.GetContentByID(ctx, page.ContentID)
	if err != nil {
		slog.Error("get content failed", "error", err)
		return echo.ErrInternalServerError
	}
	owns, err := h.teacherOwnsSubject(c, content.SubjectID)
	if err != nil {
		slog.Error("check teacher ownership failed", "error", err)
		return echo.ErrInternalServerError
	}
	if !owns {
		return echo.NewHTTPError(http.StatusForbidden, "not assigned to this subject")
	}

	var req patchPageRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if strings.TrimSpace(req.Name) == "" {
		req.Name = page.Name
	}
	if req.Position == 0 {
		req.Position = page.Position
	}

	updated, err := h.q.UpdatePage(ctx, id, strings.TrimSpace(req.Name), req.Position)
	if err != nil {
		slog.Error("update page failed", "error", err)
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusOK, map[string]any{"page": toPageResponse(updated, nil)})
}

func (h *contentHandlers) deletePage(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid page id")
	}

	page, err := h.q.GetPageByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "page not found")
		}
		slog.Error("get page failed", "error", err)
		return echo.ErrInternalServerError
	}

	content, err := h.q.GetContentByID(ctx, page.ContentID)
	if err != nil {
		slog.Error("get content failed", "error", err)
		return echo.ErrInternalServerError
	}
	owns, err := h.teacherOwnsSubject(c, content.SubjectID)
	if err != nil {
		slog.Error("check teacher ownership failed", "error", err)
		return echo.ErrInternalServerError
	}
	if !owns {
		return echo.NewHTTPError(http.StatusForbidden, "not assigned to this subject")
	}

	if err := h.q.DeletePage(ctx, id); err != nil {
		slog.Error("delete page failed", "error", err)
		return echo.ErrInternalServerError
	}
	return c.NoContent(http.StatusNoContent)
}

// ─── Blocks handler ──────────────────────────────────────────────────────────

type blockInput struct {
	Type     string          `json:"type"`
	Position int32           `json:"position"`
	Data     json.RawMessage `json:"data"`
}

type replaceBlocksRequest struct {
	Blocks []blockInput `json:"blocks"`
}

func (h *contentHandlers) replaceBlocks(c echo.Context) error {
	ctx := c.Request().Context()
	pageID, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid page id")
	}

	page, err := h.q.GetPageByID(ctx, pageID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "page not found")
		}
		slog.Error("get page failed", "error", err)
		return echo.ErrInternalServerError
	}

	content, err := h.q.GetContentByID(ctx, page.ContentID)
	if err != nil {
		slog.Error("get content failed", "error", err)
		return echo.ErrInternalServerError
	}
	owns, err := h.teacherOwnsSubject(c, content.SubjectID)
	if err != nil {
		slog.Error("check teacher ownership failed", "error", err)
		return echo.ErrInternalServerError
	}
	if !owns {
		return echo.NewHTTPError(http.StatusForbidden, "not assigned to this subject")
	}

	var blocks []blockInput

	ct := c.Request().Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/") {
		// Multipart: blocks field + optional file_<index> parts
		if err := c.Request().ParseMultipartForm(32 << 20); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid multipart form")
		}
		blocksJSON := c.Request().FormValue("blocks")
		if err := json.Unmarshal([]byte(blocksJSON), &blocks); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid blocks JSON")
		}
		// Process file uploads for each block
		for i := range blocks {
			fileKey := fmt.Sprintf("file_%d", i)
			file, header, ferr := c.Request().FormFile(fileKey)
			if ferr != nil {
				continue // no file for this block
			}
			defer file.Close()
			ext := mediaExt(header.Filename, header.Header.Get("Content-Type"))
			storageKey, kerr := newStorageKey(ext)
			if kerr != nil {
				return echo.ErrInternalServerError
			}
			if perr := h.storage.Put(ctx, storageKey, file, header.Size); perr != nil {
				slog.Error("store block file failed", "error", perr)
				return echo.ErrInternalServerError
			}
			// Inject storage_key into block data
			var dataMap map[string]any
			if len(blocks[i].Data) > 0 {
				_ = json.Unmarshal(blocks[i].Data, &dataMap)
			}
			if dataMap == nil {
				dataMap = make(map[string]any)
			}
			dataMap["storage_key"] = storageKey
			if newData, merr := json.Marshal(dataMap); merr == nil {
				blocks[i].Data = newData
			}
		}
	} else {
		// Plain JSON
		var req replaceBlocksRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}
		blocks = req.Blocks
	}

	// Validate each block
	for i, b := range blocks {
		if err := validateBlockData(b.Type, b.Data); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("block[%d] invalid: %s", i, err.Error()))
		}
	}

	// Delete all existing blocks, then insert new ones
	if err := h.q.DeleteBlocksByPage(ctx, pageID); err != nil {
		slog.Error("delete blocks failed", "error", err)
		return echo.ErrInternalServerError
	}

	result := make([]db.Block, 0, len(blocks))
	for _, b := range blocks {
		data := []byte(b.Data)
		if len(data) == 0 {
			data = []byte("{}")
		}
		created, cerr := h.q.CreateBlock(ctx, pageID, b.Type, b.Position, data)
		if cerr != nil {
			slog.Error("create block failed", "error", cerr)
			return echo.ErrInternalServerError
		}
		result = append(result, created)
	}

	resp := make([]blockResponse, len(result))
	for i, b := range result {
		resp[i] = toBlockResponse(b)
	}
	return c.JSON(http.StatusOK, map[string]any{"blocks": resp})
}

// ─── Block validation ────────────────────────────────────────────────────────

func validateBlockData(blockType string, rawData json.RawMessage) error {
	if len(rawData) == 0 {
		rawData = json.RawMessage("{}")
	}
	switch blockType {
	case "text", "article":
		var d struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(rawData, &d); err != nil {
			return fmt.Errorf("invalid data: %w", err)
		}
		if strings.TrimSpace(d.Content) == "" {
			return fmt.Errorf("content is required")
		}
	case "speech":
		var d struct {
			SourceType string `json:"source_type"`
			AudioURL   string `json:"audio_url"`
			StorageKey string `json:"storage_key"`
		}
		if err := json.Unmarshal(rawData, &d); err != nil {
			return fmt.Errorf("invalid data: %w", err)
		}
		validSrc := map[string]bool{"tts": true, "stt": true, "upload": true, "recorded": true}
		if !validSrc[d.SourceType] {
			return fmt.Errorf("source_type must be one of: tts, stt, upload, recorded")
		}
		if (d.SourceType == "upload" || d.SourceType == "recorded") && d.AudioURL == "" && d.StorageKey == "" {
			return fmt.Errorf("audio_url or storage_key required for source_type %s", d.SourceType)
		}
	case "flash_card":
		var d struct {
			Front           string `json:"front"`
			Back            string `json:"back"`
			InteractionMode string `json:"interaction_mode"`
		}
		if err := json.Unmarshal(rawData, &d); err != nil {
			return fmt.Errorf("invalid data: %w", err)
		}
		if strings.TrimSpace(d.Front) == "" {
			return fmt.Errorf("front is required")
		}
		if strings.TrimSpace(d.Back) == "" {
			return fmt.Errorf("back is required")
		}
		if strings.TrimSpace(d.InteractionMode) == "" {
			return fmt.Errorf("interaction_mode is required")
		}
	case "image":
		var d struct {
			SourceType  string `json:"source_type"`
			StorageKey  string `json:"storage_key"`
			ExternalURL string `json:"external_url"`
		}
		if err := json.Unmarshal(rawData, &d); err != nil {
			return fmt.Errorf("invalid data: %w", err)
		}
		validSrc := map[string]bool{"url": true, "upload": true, "ai": true}
		if !validSrc[d.SourceType] {
			return fmt.Errorf("source_type must be one of: url, upload, ai")
		}
		if d.SourceType == "url" && strings.TrimSpace(d.ExternalURL) == "" {
			return fmt.Errorf("external_url required for source_type url")
		}
		if d.SourceType == "upload" && strings.TrimSpace(d.StorageKey) == "" {
			return fmt.Errorf("storage_key required for source_type upload")
		}
	case "diagram":
		var d struct {
			DiagramJSON json.RawMessage `json:"diagram_json"`
		}
		if err := json.Unmarshal(rawData, &d); err != nil {
			return fmt.Errorf("invalid data: %w", err)
		}
		if len(d.DiagramJSON) == 0 || string(d.DiagramJSON) == "null" {
			return fmt.Errorf("diagram_json is required")
		}
	case "video":
		var d struct {
			SourceType string `json:"source_type"`
			EmbedURL   string `json:"embed_url"`
			StorageKey string `json:"storage_key"`
		}
		if err := json.Unmarshal(rawData, &d); err != nil {
			return fmt.Errorf("invalid data: %w", err)
		}
		validSrc := map[string]bool{"embed": true, "upload": true, "ai": true}
		if !validSrc[d.SourceType] {
			return fmt.Errorf("source_type must be one of: embed, upload, ai")
		}
		if d.SourceType == "embed" && strings.TrimSpace(d.EmbedURL) == "" {
			return fmt.Errorf("embed_url required for source_type embed")
		}
		if d.SourceType == "upload" && strings.TrimSpace(d.StorageKey) == "" {
			return fmt.Errorf("storage_key required for source_type upload")
		}
	case "translation":
		var d struct {
			SourceLanguage string            `json:"source_language"`
			Source         string            `json:"source"`
			Translations   []json.RawMessage `json:"translations"`
		}
		if err := json.Unmarshal(rawData, &d); err != nil {
			return fmt.Errorf("invalid data: %w", err)
		}
		if strings.TrimSpace(d.SourceLanguage) == "" {
			return fmt.Errorf("source_language is required")
		}
		if strings.TrimSpace(d.Source) == "" {
			return fmt.Errorf("source is required")
		}
		if len(d.Translations) == 0 {
			return fmt.Errorf("translations must be non-empty")
		}
	default:
		return fmt.Errorf("unknown block type: %s", blockType)
	}
	return nil
}

// ─── Upload serving ──────────────────────────────────────────────────────────

// ServeUploadHandler returns an Echo handler that streams files from storage.
func ServeUploadHandler(store storage.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		key := c.Param("key")
		if key == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "missing key")
		}
		// Strip any leading slash from wildcard capture
		key = strings.TrimPrefix(key, "/")

		rc, err := store.Get(c.Request().Context(), key)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "file not found")
		}
		defer rc.Close()

		ct := mime.TypeByExtension("." + extensionFromKey(key))
		if ct == "" {
			ct = "application/octet-stream"
		}
		c.Response().Header().Set("Cache-Control", "private, max-age=3600")
		return c.Stream(http.StatusOK, ct, rc)
	}
}

// ─── MIME helpers ────────────────────────────────────────────────────────────

// extensionFromKey returns the file extension (without dot) from a storage key.
func extensionFromKey(key string) string {
	dot := strings.LastIndex(key, ".")
	if dot < 0 {
		return ""
	}
	return key[dot+1:]
}

// mediaExt returns a file extension (with dot) derived from the filename or
// the Content-Type header, whichever is more reliable.
func mediaExt(filename, contentType string) string {
	if filename != "" {
		dot := strings.LastIndex(filename, ".")
		if dot >= 0 {
			return filename[dot:]
		}
	}
	if contentType != "" {
		mt, _, _ := mime.ParseMediaType(contentType)
		switch mt {
		case "image/jpeg":
			return ".jpg"
		case "image/png":
			return ".png"
		case "image/gif":
			return ".gif"
		case "image/webp":
			return ".webp"
		case "audio/mpeg":
			return ".mp3"
		case "audio/ogg":
			return ".ogg"
		case "audio/wav":
			return ".wav"
		case "video/mp4":
			return ".mp4"
		case "video/webm":
			return ".webm"
		}
	}
	return ""
}
