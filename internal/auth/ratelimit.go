package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"

	"librarie/internal/db"
)

const (
	bucketCapacity = int32(5)
	refillPerMin   = float64(1)
)

// PasswordFingerprint computes a dedup key for a (username, password) pair.
// Used to skip token consumption when the same wrong password is retried.
func PasswordFingerprint(username, password string) string {
	sum := sha256.Sum256([]byte(username + ":" + password))
	return hex.EncodeToString(sum[:])
}

// RateLimiter implements a token-bucket rate limiter backed by the DB.
type RateLimiter struct {
	q *db.Queries
}

// NewRateLimiter creates a RateLimiter using the provided query set.
func NewRateLimiter(q *db.Queries) *RateLimiter {
	return &RateLimiter{q: q}
}

// consume tries to take one token from the bucket identified by (scopeType, scopeKey).
// Returns (true, 0) on success or (false, retryAfterSecs) when the bucket is empty.
func (r *RateLimiter) consume(ctx context.Context, scopeType, scopeKey string) (ok bool, retryAfter int, err error) {
	row, dbErr := r.q.GetLoginRateLimit(ctx, scopeType, scopeKey)

	var currentTokens int32
	var lastUpdate time.Time

	if dbErr != nil {
		if dbErr == pgx.ErrNoRows {
			currentTokens = bucketCapacity
			lastUpdate = time.Now()
		} else {
			return false, 0, dbErr
		}
	} else {
		lastUpdate = row.UpdatedAt.Time
		elapsed := time.Since(lastUpdate)
		refilled := int32(math.Floor(elapsed.Minutes() * refillPerMin))
		currentTokens = min32(row.Tokens+refilled, bucketCapacity)
	}

	if currentTokens <= 0 {
		retryAfter = int(math.Ceil(60 / refillPerMin))
		return false, retryAfter, nil
	}

	newTokens := currentTokens - 1
	if _, err = r.q.UpsertLoginRateLimit(ctx, scopeType, scopeKey, newTokens); err != nil {
		return false, 0, err
	}
	return true, 0, nil
}

// loginBody is a minimal struct for peeking username/password in middleware.
type loginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRateLimitMiddleware applies IP-scoped and username-scoped token-bucket limits.
// It also deduplicates repeated identical failed attempts to avoid burning extra tokens.
func (r *RateLimiter) LoginRateLimitMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			ip := c.RealIP()

			// Buffer body so it can be re-read by the handler.
			rawBody, err := io.ReadAll(c.Request().Body)
			if err != nil {
				return echo.ErrInternalServerError
			}
			c.Request().Body = io.NopCloser(bytes.NewReader(rawBody))

			// Parse username + password for dedup check.
			var body loginBody
			_ = json.Unmarshal(rawBody, &body)

			// Check dedup: if same (username, password) was already attempted
			// recently, skip token consumption (don't burn another token).
			if body.Username != "" && body.Password != "" {
				fingerprint := PasswordFingerprint(body.Username, body.Password)
				attempts, _ := r.q.ListRecentLoginAttemptsByUsername(ctx, body.Username)
				for _, a := range attempts {
					if a.PasswordFingerprint != nil && *a.PasswordFingerprint == fingerprint && !a.Success {
						// Duplicate trial — pass through without consuming tokens.
						return next(c)
					}
				}
			}

			// IP-scoped bucket.
			ok, retryAfter, err := r.consume(ctx, "ip", ip)
			if err != nil {
				return echo.ErrInternalServerError
			}
			if !ok {
				c.Response().Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				return echo.NewHTTPError(http.StatusTooManyRequests, "too many requests")
			}

			// Username-scoped bucket.
			if body.Username != "" {
				ok, retryAfter, err = r.consume(ctx, "username", body.Username)
				if err != nil {
					return echo.ErrInternalServerError
				}
				if !ok {
					c.Response().Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
					return echo.NewHTTPError(http.StatusTooManyRequests, "too many requests")
				}
			}

			return next(c)
		}
	}
}

// IPRateLimitMiddleware applies IP-scoped rate limiting only (for passkey authenticate).
func (r *RateLimiter) IPRateLimitMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			ip := c.RealIP()

			ok, retryAfter, err := r.consume(ctx, "ip", ip)
			if err != nil {
				return echo.ErrInternalServerError
			}
			if !ok {
				c.Response().Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				return echo.NewHTTPError(http.StatusTooManyRequests, "too many requests")
			}
			return next(c)
		}
	}
}

func min32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
