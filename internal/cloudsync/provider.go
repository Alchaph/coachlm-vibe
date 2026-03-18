package cloudsync

import (
	"context"
	"time"
)

// CloudProvider defines the interface for cloud storage backends.
// Implementations must be safe for concurrent use.
type CloudProvider interface {
	// Upload stores data at the given key, overwriting any existing content.
	Upload(ctx context.Context, key string, data []byte) error

	// Download retrieves data stored at the given key.
	// Returns ErrNotFound if the key does not exist.
	Download(ctx context.Context, key string) ([]byte, error)

	// LastModified returns the last modification time for the given key.
	// Returns ErrNotFound if the key does not exist.
	LastModified(ctx context.Context, key string) (time.Time, error)

	// Name returns a human-readable provider name (e.g. "Google Drive", "S3").
	Name() string
}
