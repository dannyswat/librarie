package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

// safeKeyRe matches allowed storage key characters: hex digits and dots.
// Keys are generated as <32-hex-chars>.<ext> so this is sufficient.
var safeKeyRe = regexp.MustCompile(`^[a-zA-Z0-9\-_.]+$`)

// LocalStorage implements Storage using the local file system.
type LocalStorage struct {
	basePath string
}

// NewLocalStorage returns a LocalStorage rooted at basePath.
// The directory is created if it does not already exist.
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	abs, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("storage: resolve base path: %w", err)
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, fmt.Errorf("storage: create base directory: %w", err)
	}
	return &LocalStorage{basePath: abs}, nil
}

func (s *LocalStorage) safePath(key string) (string, error) {
	if !safeKeyRe.MatchString(key) {
		return "", fmt.Errorf("storage: invalid key %q", key)
	}
	return filepath.Join(s.basePath, key), nil
}

func (s *LocalStorage) Put(ctx context.Context, key string, r io.Reader, size int64) error {
	p, err := s.safePath(key)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
	if err != nil {
		return fmt.Errorf("storage: create file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("storage: write file: %w", err)
	}
	return nil
}

func (s *LocalStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	p, err := s.safePath(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("storage: key not found: %s", key)
		}
		return nil, fmt.Errorf("storage: open file: %w", err)
	}
	return f, nil
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	p, err := s.safePath(key)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("storage: delete file: %w", err)
	}
	return nil
}

// URL returns the backend URL path for the given key.
func (s *LocalStorage) URL(key string) string {
	return "/uploads/" + key
}
