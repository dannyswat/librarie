package auth

import (
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"librarie/internal/db"
)

// PasskeyHandlers holds all WebAuthn handler functions.
type PasskeyHandlers struct {
	wa *webauthn.WebAuthn
	q  *db.Queries
}

// NewPasskeyHandlers creates a PasskeyHandlers instance.
func NewPasskeyHandlers(wa *webauthn.WebAuthn, q *db.Queries) *PasskeyHandlers {
	return &PasskeyHandlers{wa: wa, q: q}
}

// ── Registration ─────────────────────────────────────────────────────────────

// RegisterBegin handles POST /api/v1/auth/passkey/register/begin.
// Requires an authenticated session.
func (h *PasskeyHandlers) RegisterBegin(c echo.Context) error {
	user := UserFromContext(c)
	ctx := c.Request().Context()

	passkeys, err := h.q.ListPasskeysByUserID(ctx, user.ID)
	if err != nil {
		slog.Error("passkey register begin: list passkeys failed", "error", err)
		return echo.ErrInternalServerError
	}

	creds := make([]webauthn.Credential, len(passkeys))
	for i, p := range passkeys {
		creds[i] = dbPasskeyToCredential(p)
	}

	waUser := &WebAuthnUser{User: user, Credentials: creds}

	options, sessionData, err := h.wa.BeginRegistration(waUser)
	if err != nil {
		slog.Error("passkey register begin: BeginRegistration failed", "error", err)
		return echo.ErrInternalServerError
	}

	challengeKey, err := randomChallengeKey()
	if err != nil {
		return echo.ErrInternalServerError
	}
	globalChallengeStore.Set(challengeKey, sessionData)
	setChallengeCookie(c, challengeKey)

	return c.JSON(http.StatusOK, options)
}

// RegisterComplete handles POST /api/v1/auth/passkey/register/complete.
// Requires an authenticated session.
func (h *PasskeyHandlers) RegisterComplete(c echo.Context) error {
	user := UserFromContext(c)
	ctx := c.Request().Context()

	challengeKey, err := getChallengeCookie(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "missing challenge cookie")
	}

	sessionData, ok := globalChallengeStore.Get(challengeKey)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "challenge expired or not found")
	}

	passkeys, err := h.q.ListPasskeysByUserID(ctx, user.ID)
	if err != nil {
		slog.Error("passkey register complete: list passkeys failed", "error", err)
		return echo.ErrInternalServerError
	}

	creds := make([]webauthn.Credential, len(passkeys))
	for i, p := range passkeys {
		creds[i] = dbPasskeyToCredential(p)
	}

	waUser := &WebAuthnUser{User: user, Credentials: creds}

	cred, err := h.wa.FinishRegistration(waUser, *sessionData, c.Request())
	if err != nil {
		slog.Warn("passkey register complete: FinishRegistration failed", "error", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "passkey registration failed")
	}

	fields := credentialToPasskeyFields(user.ID, cred)
	if _, err := h.q.CreatePasskey(ctx, fields.UserID, fields.CredentialID, fields.PublicKey, fields.SignCount); err != nil {
		slog.Error("passkey register complete: store passkey failed", "error", err)
		return echo.ErrInternalServerError
	}

	clearChallengeCookie(c)
	return c.JSON(http.StatusCreated, map[string]string{"status": "ok"})
}

// ── Authentication ────────────────────────────────────────────────────────────

// AuthenticateBegin handles POST /api/v1/auth/passkey/authenticate/begin.
func (h *PasskeyHandlers) AuthenticateBegin(c echo.Context) error {
	options, sessionData, err := h.wa.BeginDiscoverableLogin()
	if err != nil {
		slog.Error("passkey auth begin: BeginDiscoverableLogin failed", "error", err)
		return echo.ErrInternalServerError
	}

	challengeKey, err := randomChallengeKey()
	if err != nil {
		return echo.ErrInternalServerError
	}
	globalChallengeStore.Set(challengeKey, sessionData)
	setChallengeCookie(c, challengeKey)

	return c.JSON(http.StatusOK, options)
}

// AuthenticateComplete handles POST /api/v1/auth/passkey/authenticate/complete.
func (h *PasskeyHandlers) AuthenticateComplete(c echo.Context) error {
	ctx := c.Request().Context()

	challengeKey, err := getChallengeCookie(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "missing challenge cookie")
	}

	sessionData, ok := globalChallengeStore.Get(challengeKey)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "challenge expired or not found")
	}

	var lookedUpUser *WebAuthnUser

	cred, err := h.wa.FinishDiscoverableLogin(
		func(rawID, userHandle []byte) (webauthn.User, error) {
			// Look up passkey by credential ID.
			passkey, err := h.q.GetPasskeyByCredentialID(ctx, rawID)
			if err != nil {
				return nil, err
			}
			user, err := h.q.GetUserByID(ctx, passkey.UserID)
			if err != nil {
				return nil, err
			}
			passkeys, err := h.q.ListPasskeysByUserID(ctx, user.ID)
			if err != nil {
				return nil, err
			}
			creds := make([]webauthn.Credential, len(passkeys))
			for i, p := range passkeys {
				creds[i] = dbPasskeyToCredential(p)
			}
			waUser := &WebAuthnUser{User: user, Credentials: creds}
			lookedUpUser = waUser
			return waUser, nil
		},
		*sessionData,
		c.Request(),
	)
	if err != nil {
		slog.Warn("passkey auth complete: FinishDiscoverableLogin failed", "error", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "passkey authentication failed")
	}

	if lookedUpUser == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "passkey authentication failed")
	}

	// Update sign count.
	passkey, err := h.q.GetPasskeyByCredentialID(ctx, cred.ID)
	if err != nil && err != pgx.ErrNoRows {
		slog.Error("passkey auth complete: sign count lookup failed", "error", err)
	}
	if err == nil {
		if _, err := h.q.UpdatePasskeySignCount(ctx, passkey.ID, int64(cred.Authenticator.SignCount)); err != nil {
			slog.Warn("passkey auth complete: sign count update failed", "error", err)
		}
	}

	// Create session.
	token, err := GenerateToken()
	if err != nil {
		slog.Error("passkey auth complete: token generation failed", "error", err)
		return echo.ErrInternalServerError
	}

	expiresAt := pgtype.Timestamptz{Time: time.Now().Add(SessionDuration), Valid: true}
	if _, err := h.q.CreateSession(ctx, lookedUpUser.User.ID, HashToken(token), expiresAt); err != nil {
		slog.Error("passkey auth complete: session creation failed", "error", err)
		return echo.ErrInternalServerError
	}

	clearChallengeCookie(c)
	SetSessionCookie(c, token)
	return c.JSON(http.StatusOK, toUserResponse(lookedUpUser.User))
}

// ── Cookie helpers ────────────────────────────────────────────────────────────

func randomChallengeKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func setChallengeCookie(c echo.Context, key string) {
	c.SetCookie(&http.Cookie{
		Name:     challengeCookieName,
		Value:    key,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   300, // 5 minutes
	})
}

func getChallengeCookie(c echo.Context) (string, error) {
	cookie, err := c.Cookie(challengeCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func clearChallengeCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     challengeCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}
