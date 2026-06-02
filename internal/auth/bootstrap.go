package auth

import (
	"context"
	"crypto/rand"
	"log/slog"

	"librarie/internal/config"
	"librarie/internal/db"
)

// BootstrapAdmins ensures all usernames listed in config are present in the DB
// with role=admin. New accounts receive a random unusable password hash.
func BootstrapAdmins(ctx context.Context, q *db.Queries, cfg *config.Config) {
	for _, username := range cfg.AdminBootstrapUsers {
		user, err := q.GetUserByUsername(ctx, username)
		if err == nil {
			// User exists — elevate to admin if needed.
			if user.Role != "admin" {
				if _, err := q.UpdateUserRole(ctx, user.ID, "admin"); err != nil {
					slog.Error("bootstrap: failed to elevate user to admin", "username", username, "error", err)
				} else {
					slog.Info("bootstrap: elevated existing user to admin", "username", username)
				}
			}
			continue
		}

		// User does not exist — create with an unusable random password hash.
		unusableHash, hashErr := randomUnusableHash()
		if hashErr != nil {
			slog.Error("bootstrap: failed to generate unusable hash", "error", hashErr)
			continue
		}

		if _, err := q.CreateUser(ctx, username, username+"@bootstrap.local", unusableHash, "admin"); err != nil {
			slog.Error("bootstrap: failed to create admin user", "username", username, "error", err)
		} else {
			slog.Info("bootstrap: created admin user", "username", username)
		}
	}
}

// randomUnusableHash generates a bcrypt-like unusable hash from random bytes.
// The resulting string cannot be matched by any real password.
func randomUnusableHash() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// Prefix with '!' so any comparison will fail deterministically.
	return "!invalid:" + HashToken(string(b)), nil
}
