package user

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"librarie/internal/auth"
	"librarie/internal/db"
)

const invitationTTL = 7 * 24 * time.Hour

var defaultCapabilities = []string{
	"generate_text",
	"generate_audio",
	"generate_image",
	"generate_video",
	"transcribe_audio",
	"synthesize_speech",
}

type adminHandlers struct {
	q      *db.Queries
	cipher *credentialCipher
}

type subjectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type teacherResponse struct {
	ID        string            `json:"id"`
	Username  string            `json:"username"`
	Email     string            `json:"email"`
	Role      string            `json:"role"`
	CreatedAt string            `json:"created_at"`
	Subjects  []subjectResponse `json:"subjects"`
}

type inviteTeacherRequest struct {
	Email string `json:"email"`
}

type replaceTeacherSubjectsRequest struct {
	SubjectIDs []string `json:"subject_ids"`
}

type upsertCapabilityRequest struct {
	Model       string         `json:"model"`
	IsEnabled   bool           `json:"is_enabled"`
	Credentials map[string]any `json:"credentials"`
}

type capabilityResponse struct {
	ProviderKey string `json:"provider_key"`
	Capability  string `json:"capability"`
	Model       string `json:"model"`
	IsEnabled   bool   `json:"is_enabled"`
}

func (h *adminHandlers) listTeachers(c echo.Context) error {
	ctx := c.Request().Context()

	teachers, err := h.q.ListUsersByRole(ctx, "teacher")
	if err != nil {
		slog.Error("list teachers failed", "error", err)
		return echo.ErrInternalServerError
	}

	resp := make([]teacherResponse, 0, len(teachers))
	for _, teacher := range teachers {
		subjects, err := h.q.ListSubjectsByTeacher(ctx, teacher.ID)
		if err != nil {
			slog.Error("list teacher subjects failed", "teacher_id", teacher.ID.String(), "error", err)
			return echo.ErrInternalServerError
		}
		resp = append(resp, toTeacherResponse(teacher, subjects))
	}

	return c.JSON(http.StatusOK, map[string]any{"teachers": resp})
}

func (h *adminHandlers) inviteTeacher(c echo.Context) error {
	caller := auth.UserFromContext(c)

	var req inviteTeacherRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if strings.TrimSpace(req.Email) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}

	token, err := generateInviteToken()
	if err != nil {
		slog.Error("invite teacher token generation failed", "error", err)
		return echo.ErrInternalServerError
	}

	expiresAt := pgtype.Timestamptz{Time: time.Now().Add(invitationTTL), Valid: true}
	inv, err := h.q.CreateInvitation(c.Request().Context(), db.CreateInvitationParams{
		Email:     req.Email,
		Role:      "teacher",
		TokenHash: auth.HashToken(token),
		InvitedBy: caller.ID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		slog.Error("invite teacher db insert failed", "error", err)
		return echo.ErrInternalServerError
	}

	slog.Info("teacher invitation created",
		"id", inv.ID.String(),
		"email", inv.Email,
		"token", token,
	)

	return c.JSON(http.StatusCreated, map[string]string{
		"id":         inv.ID.String(),
		"email":      inv.Email,
		"role":       inv.Role,
		"expires_at": inv.ExpiresAt.Time.UTC().Format(time.RFC3339),
	})
}

func (h *adminHandlers) replaceTeacherSubjects(c echo.Context) error {
	teacherID, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid teacher id")
	}

	var req replaceTeacherSubjectsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	ctx := c.Request().Context()

	teacher, err := h.q.GetUserByID(ctx, teacherID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "teacher not found")
		}
		slog.Error("lookup teacher failed", "error", err)
		return echo.ErrInternalServerError
	}
	if teacher.Role != "teacher" {
		return echo.NewHTTPError(http.StatusBadRequest, "user is not a teacher")
	}

	subjectIDs := make([]pgtype.UUID, 0, len(req.SubjectIDs))
	seen := make(map[string]struct{}, len(req.SubjectIDs))
	for _, rawID := range req.SubjectIDs {
		trimmed := strings.TrimSpace(rawID)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}

		subjectID, err := parseUUIDParam(trimmed)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid subject id: "+trimmed)
		}
		if _, err := h.q.GetSubjectByID(ctx, subjectID); err != nil {
			if err == pgx.ErrNoRows {
				return echo.NewHTTPError(http.StatusBadRequest, "subject not found: "+trimmed)
			}
			slog.Error("lookup subject failed", "subject_id", trimmed, "error", err)
			return echo.ErrInternalServerError
		}

		subjectIDs = append(subjectIDs, subjectID)
	}

	if err := h.q.DeleteTeacherSubjectAssignments(ctx, teacherID); err != nil {
		slog.Error("delete teacher subject assignments failed", "teacher_id", teacherID.String(), "error", err)
		return echo.ErrInternalServerError
	}

	caller := auth.UserFromContext(c)
	for _, subjectID := range subjectIDs {
		if err := h.q.AssignTeacherToSubject(ctx, teacherID, subjectID, caller.ID); err != nil {
			slog.Error("assign teacher to subject failed", "teacher_id", teacherID.String(), "subject_id", subjectID.String(), "error", err)
			return echo.ErrInternalServerError
		}
	}

	subjects, err := h.q.ListSubjectsByTeacher(ctx, teacherID)
	if err != nil {
		slog.Error("list teacher subjects failed", "teacher_id", teacherID.String(), "error", err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, map[string]any{
		"teacher_id": teacherID.String(),
		"subjects":   toSubjectResponses(subjects),
	})
}

func (h *adminHandlers) listProviderCapabilities(c echo.Context) error {
	providerKey := strings.TrimSpace(c.Param("provider_key"))
	if providerKey == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "provider_key is required")
	}

	rows, err := h.q.ListAIProviderConfigsByProvider(c.Request().Context(), providerKey)
	if err != nil {
		slog.Error("list provider capabilities failed", "provider_key", providerKey, "error", err)
		return echo.ErrInternalServerError
	}

	byCapability := make(map[string]db.AiProviderConfig, len(rows))
	for _, row := range rows {
		byCapability[row.Capability] = row
	}

	caps := make([]capabilityResponse, 0, len(defaultCapabilities))
	for _, capability := range defaultCapabilities {
		if row, ok := byCapability[capability]; ok {
			caps = append(caps, toCapabilityResponse(row))
			continue
		}
		caps = append(caps, capabilityResponse{
			ProviderKey: providerKey,
			Capability:  capability,
			Model:       "",
			IsEnabled:   false,
		})
	}

	for capability, row := range byCapability {
		if isKnownCapability(capability) {
			continue
		}
		caps = append(caps, toCapabilityResponse(row))
	}

	sort.Slice(caps, func(i, j int) bool {
		return caps[i].Capability < caps[j].Capability
	})

	return c.JSON(http.StatusOK, map[string]any{
		"provider_key": providerKey,
		"capabilities": caps,
	})
}

func (h *adminHandlers) upsertProviderCapability(c echo.Context) error {
	providerKey := strings.TrimSpace(c.Param("provider_key"))
	capability := strings.TrimSpace(c.Param("capability"))
	if providerKey == "" || capability == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "provider_key and capability are required")
	}
	if !isKnownCapability(capability) {
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported capability")
	}

	var req upsertCapabilityRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if strings.TrimSpace(req.Model) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "model is required")
	}

	credentialBytes, err := json.Marshal(req.Credentials)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid credentials payload")
	}
	encrypted, err := h.cipher.encrypt(credentialBytes)
	if err != nil {
		slog.Error("encrypt provider credentials failed", "provider_key", providerKey, "capability", capability, "error", err)
		return echo.ErrInternalServerError
	}

	updated, err := h.q.UpsertAIProviderConfig(c.Request().Context(), db.UpsertAIProviderConfigParams{
		ProviderKey:          providerKey,
		Capability:           capability,
		EncryptedCredentials: encrypted,
		Model:                strings.TrimSpace(req.Model),
		IsEnabled:            req.IsEnabled,
	})
	if err != nil {
		slog.Error("upsert provider capability failed", "provider_key", providerKey, "capability", capability, "error", err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, toCapabilityResponse(updated))
}

func toTeacherResponse(teacher db.User, subjects []db.Subject) teacherResponse {
	return teacherResponse{
		ID:        teacher.ID.String(),
		Username:  teacher.Username,
		Email:     teacher.Email,
		Role:      teacher.Role,
		CreatedAt: teacher.CreatedAt.Time.UTC().Format(time.RFC3339),
		Subjects:  toSubjectResponses(subjects),
	}
}

func toSubjectResponses(subjects []db.Subject) []subjectResponse {
	out := make([]subjectResponse, 0, len(subjects))
	for _, subject := range subjects {
		out = append(out, subjectResponse{
			ID:          subject.ID.String(),
			Name:        subject.Name,
			Description: subject.Description,
		})
	}
	return out
}

func toCapabilityResponse(row db.AiProviderConfig) capabilityResponse {
	return capabilityResponse{
		ProviderKey: row.ProviderKey,
		Capability:  row.Capability,
		Model:       row.Model,
		IsEnabled:   row.IsEnabled,
	}
}

func isKnownCapability(capability string) bool {
	for _, c := range defaultCapabilities {
		if c == capability {
			return true
		}
	}
	return false
}

func parseUUIDParam(value string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := id.Scan(value)
	return id, err
}

func generateInviteToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
