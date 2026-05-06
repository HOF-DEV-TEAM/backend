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

// Ensure CloudinaryStorage implements the Storage interface
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
func (c *CloudinaryStorage) Upload(_ context.Context, fh *multipart.FileHeader, key string) (string, error) {
	f, err := fh.Open()
	if err != nil {
		return "", shared.ErrInvalidInput{Message: "failed to open upload file"}
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			c.log.Warn("closing upload file", zap.Error(cerr))
		}
	}()

	// Prepare upload parameters
	uploadParams := uploader.UploadParams{
		PublicID:     key,
		ResourceType: "auto", // Let Cloudinary detect the resource type
		Folder:       "uploads",
	}

	// If upload preset is configured, use it
	if c.cfg.UploadPreset != "" {
		uploadParams.UploadPreset = c.cfg.UploadPreset
	}

	// Upload to Cloudinary using the file reader directly
	result, err := c.cld.Upload.Upload(context.Background(), f, uploadParams)
	if err != nil {
		return "", fmt.Errorf("cloudinary upload: %w", err)
	}

	url := result.SecureURL
	c.log.Info("file uploaded to Cloudinary", zap.String("url", url))
	return url, nil
}
