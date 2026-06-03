package config

import (
	"bufio"
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
	// Attempt to load variables from a local .env file (do not override existing env vars)
	loadDotEnv()

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

// loadDotEnv reads a .env file in the current working directory and sets any
// variables that are not already present in the environment. It's intentionally
// small and permissive — comments and blank lines are ignored, and values may
// be wrapped in single or double quotes.
func loadDotEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split at first '='
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])

		// Strip surrounding quotes if present
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}

		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}

	if err := scanner.Err(); err != nil {
		// Ignore scanning errors — failure to parse .env should not stop the app.
		_ = err
	}
}
