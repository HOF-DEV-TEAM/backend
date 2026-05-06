// Package storage provides a common interface for file storage backends.
package storage

import (
	"context"
	"mime/multipart"
)

// Storage defines the interface for file storage implementations.
type Storage interface {
	// Upload stores a file and returns the public URL.
	Upload(ctx context.Context, fh *multipart.FileHeader, key string) (string, error)
}
