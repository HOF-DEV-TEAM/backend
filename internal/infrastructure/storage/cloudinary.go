// Package storage provides Cloudinary upload helpers.
package storage

import (
	"context"
	"fmt"
	"mime/multipart"

	"bitbucket.org/hofng/hofApp/internal/domain/shared"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/config"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"go.uber.org/zap"
)

// CloudinaryStorage uploads files to Cloudinary.
type CloudinaryStorage struct {
	cld *cloudinary.Cloudinary
	cfg *config.CloudinaryConfig
	log *zap.Logger
}

// Ensure CloudinaryStorage implements the Storage interface.
var _ Storage = (*CloudinaryStorage)(nil)

// NewCloudinaryStorage connects to Cloudinary and returns a ready-to-use CloudinaryStorage.
func NewCloudinaryStorage(cfg *config.CloudinaryConfig, log *zap.Logger) (*CloudinaryStorage, error) {
	if cfg == nil {
		cfg = &config.CloudinaryConfig{}
	}

	cld, err := cloudinary.NewFromParams(cfg.CloudName, cfg.APIKey, cfg.APISecret)
	if err != nil {
		return nil, shared.ErrInvalidInput{Message: "invalid Cloudinary configuration"}
	}

	return &CloudinaryStorage{
		cld: cld,
		cfg: cfg,
		log: log,
	}, nil
}

// Upload stores the file in Cloudinary and returns the public URL.
func (c *CloudinaryStorage) Upload(ctx context.Context, fh *multipart.FileHeader, key string) (string, error) {
	f, err := fh.Open()
	if err != nil {
		return "", shared.ErrInvalidInput{Message: "failed to open upload file"}
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			c.log.Warn("closing upload file", zap.Error(cerr))
		}
	}()

	uploadParams := uploader.UploadParams{
		PublicID:     key,
		ResourceType: "auto",
		Folder:       "uploads",
	}

	if c.cfg.UploadPreset != "" {
		uploadParams.UploadPreset = c.cfg.UploadPreset
	}

	result, err := c.cld.Upload.Upload(ctx, f, uploadParams)
	if err != nil {
		return "", fmt.Errorf("cloudinary upload: %w", err)
	}

	url := result.SecureURL
	c.log.Info("file uploaded to Cloudinary", zap.String("url", url), zap.Int64("size", fh.Size))
	return url, nil
}

// GeneratePresignedURL is a no-op for Cloudinary as it doesn't support presigned URLs.
// Use Upload() directly or implement Cloudinary's upload widget instead.
func (c *CloudinaryStorage) GeneratePresignedURL(_ context.Context, _ string, _ string) (string, error) {
	return "", shared.ErrInvalidInput{Message: "presigned URLs not supported for Cloudinary"}
}

// GetMaxFileSize returns the Cloudinary max file size (500MB).
func (c *CloudinaryStorage) GetMaxFileSize() int64 {
	// Cloudinary free tier limit: 5MB per file, pro tier: 500MB
	// Return 500MB as default; adjust if needed
	return 524288000 // 500MB
}
