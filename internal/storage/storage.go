package storage

import (
	"context"
	"io"
)

// Storage is the abstraction for binary file persistence.
type Storage interface {
	// Put stores content from r under the given key.
	Put(ctx context.Context, key string, r io.Reader, size int64) error

	// Get retrieves the content stored at key.
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes the object stored at key (no-op if key does not exist).
	Delete(ctx context.Context, key string) error

	// URL returns the backend URL path that serves this key.
	URL(key string) string
}
