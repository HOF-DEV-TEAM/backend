// Package storage provides a factory for creating storage backends.
package storage

import (
	"fmt"

	"bitbucket.org/hofng/hofApp/internal/infrastructure/config"
	"go.uber.org/zap"
)

// NewStorage creates a storage backend based on configuration.
// It prioritizes S3 if AWS endpoint is configured, otherwise falls back to Cloudinary.
func NewStorage(cfg *config.ServerConfig, log *zap.Logger) (Storage, error) {
	// Try S3 first if AWS secret is configured
	if cfg.AWS.Secret != "" {
		log.Info("initializing S3 storage")
		s3Storage, err := NewS3Storage(&cfg.AWS, log)
		if err != nil {
			return nil, fmt.Errorf("creating S3 storage: %w", err)
		}
		return s3Storage, nil
	}

	// Fall back to Cloudinary if S3 is not configured
	if cfg.Cloudinary.CloudName != "" && cfg.Cloudinary.APIKey != "" && cfg.Cloudinary.APISecret != "" {
		log.Info("initializing Cloudinary storage (S3 not configured)")
		cloudinaryStorage, err := NewCloudinaryStorage(&cfg.Cloudinary, log)
		if err != nil {
			return nil, fmt.Errorf("creating Cloudinary storage: %w", err)
		}
		return cloudinaryStorage, nil
	}

	return nil, fmt.Errorf("no storage backend configured: please configure either AWS (AWS_ENDPOINT) or Cloudinary (CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY, CLOUDINARY_API_SECRET)")
}
