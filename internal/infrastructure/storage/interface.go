// Package storage provides a common interface for file storage backends.
package storage

import (
	"context"
	"mime/multipart"
)

// FileInfo contains metadata for uploaded files.
type FileInfo struct {
	Filename    string
	ContentType string
	Size        int64
}

// Storage defines the interface for file storage implementations.
type Storage interface {
	// Upload stores a file and returns the public URL.
	Upload(ctx context.Context, fh *multipart.FileHeader, key string) (string, error)

	// GeneratePresignedURL returns a time-limited upload URL.
	// Returns (uploadURL, error).
	GeneratePresignedURL(ctx context.Context, key string, contentType string) (string, error)

	// GetMaxFileSize returns the maximum allowed file size in bytes.
	GetMaxFileSize() int64
}
