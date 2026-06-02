package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	// Server
	Port string

	// Database
	DatabaseURL string

	// Session
	SessionSecret string

	// Admin bootstrap: comma-separated usernames that are auto-elevated to admin
	AdminBootstrapUsers []string

	// Encryption key for AI provider credentials (32 bytes, hex-encoded)
	EncryptionKey string

	// Storage
	StorageBasePath string

	// WebAuthn / Passkey
	RPDisplayName string
	RPID          string
	RPOrigin      string
}

// Load reads configuration from environment variables, applying defaults where safe.
// It returns an error if any required variable is missing.
func Load() (*Config, error) {
	cfg := &Config{
		Port:            getEnv("LIBRARIE_PORT", "8080"),
		DatabaseURL:     getEnv("LIBRARIE_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/librarie?sslmode=disable"),
		SessionSecret:   getEnv("LIBRARIE_SESSION_SECRET", ""),
		EncryptionKey:   getEnv("LIBRARIE_ENCRYPTION_KEY", ""),
		StorageBasePath: getEnv("LIBRARIE_STORAGE_PATH", "./data/uploads"),
		RPDisplayName:   getEnv("LIBRARIE_RP_DISPLAY_NAME", "Librarie"),
		RPID:            getEnv("LIBRARIE_RP_ID", "localhost"),
		RPOrigin:        getEnv("LIBRARIE_RP_ORIGIN", "http://localhost:5173"),
	}

	adminRaw := getEnv("LIBRARIE_ADMIN_USERS", "")
	if adminRaw != "" {
		for _, u := range strings.Split(adminRaw, ",") {
			u = strings.TrimSpace(u)
			if u != "" {
				cfg.AdminBootstrapUsers = append(cfg.AdminBootstrapUsers, u)
			}
		}
	}

	var missing []string
	if cfg.SessionSecret == "" {
		missing = append(missing, "LIBRARIE_SESSION_SECRET")
	}
	if cfg.EncryptionKey == "" {
		missing = append(missing, "LIBRARIE_ENCRYPTION_KEY")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("required environment variables not set: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
