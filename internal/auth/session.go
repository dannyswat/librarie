package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	sessionCookieName = "librarie_session"
	SessionDuration   = 30 * 24 * time.Hour
)

// GenerateToken creates a cryptographically random opaque session token.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken returns the SHA-256 hex digest of the token, suitable for DB storage.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// SetSessionCookie writes the HttpOnly session cookie to the response.
func SetSessionCookie(c echo.Context, token string) {
	c.SetCookie(&http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(SessionDuration.Seconds()),
	})
}

// ClearSessionCookie removes the session cookie.
func ClearSessionCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// GetSessionToken reads the raw token from the request cookie.
func GetSessionToken(c echo.Context) (string, error) {
	cookie, err := c.Cookie(sessionCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetBearerToken returns the token from Authorization: Bearer <token>.
func GetBearerToken(c echo.Context) (string, bool) {
	authz := strings.TrimSpace(c.Request().Header.Get("Authorization"))
	if authz == "" {
		return "", false
	}
	if !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		return "", false
	}
	token := strings.TrimSpace(authz[7:])
	if token == "" {
		return "", false
	}
	return token, true
}
