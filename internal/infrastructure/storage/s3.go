// Package storage provides AWS S3 upload helpers.
package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"bitbucket.org/hofng/hofApp/internal/domain/shared"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"
)

// S3Storage uploads files to AWS S3.
type S3Storage struct {
	uploader *s3manager.Uploader
	s3Client *s3.S3
	cfg      *config.AWSConfig
	log      *zap.Logger
}

// Ensure S3Storage implements the Storage interface.
var _ Storage = (*S3Storage)(nil)

// NewS3Storage connects to AWS and returns a ready-to-use S3Storage.
func NewS3Storage(cfg *config.AWSConfig, log *zap.Logger) (*S3Storage, error) {
	if cfg == nil {
		cfg = &config.AWSConfig{}
	}
	sess, err := session.NewSession(&aws.Config{
		Region:      &cfg.Region,
		Credentials: credentials.NewStaticCredentials(cfg.AccessKey, cfg.Secret, ""),
	})
	if err != nil {
		return nil, shared.ErrInvalidInput{Message: "invalid AWS configuration"}
	}

	// Configure uploader with multipart settings for optimal performance
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		// Set part size for multipart uploads (default 10MB)
		if cfg.PartSize > 0 {
			u.PartSize = cfg.PartSize
		}
		// Concurrent parts upload (default 5 concurrent parts)
		u.Concurrency = 5
	})

	return &S3Storage{
		uploader: uploader,
		s3Client: s3.New(sess),
		cfg:      cfg,
		log:      log,
	}, nil
}

// Upload stores the file in S3 under <bucketPath><key> and returns the public URL.
// Uses multipart upload for optimal performance with large files.
func (s *S3Storage) Upload(ctx context.Context, fh *multipart.FileHeader, key string) (string, error) {
	// Validate file size early
	if fh.Size > s.cfg.MaxFileSize {
		return "", shared.ErrInvalidInput{
			Field:   "file",
			Message: fmt.Sprintf("file size %d exceeds maximum of %d bytes", fh.Size, s.cfg.MaxFileSize),
		}
	}

	f, err := fh.Open()
	if err != nil {
		return "", shared.ErrInvalidInput{Message: "failed to open upload file"}
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			s.log.Warn("closing upload file", zap.Error(cerr))
		}
	}()

	objectKey := s.cfg.BucketPath + key
	contentType := fh.Header.Get("Content-Type")

	// Multipart upload is automatic for large files; s3manager handles it
	_, err = s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:      &s.cfg.Bucket,
		Key:         &objectKey,
		Body:        f,
		ContentType: &contentType,
	})
	if err != nil {
		return "", fmt.Errorf("s3 upload: %w", err)
	}

	url := fmt.Sprintf("%s%s/%s", s.cfg.BaseURL, s.cfg.Bucket, objectKey)
	s.log.Info("file uploaded to S3", zap.String("url", url), zap.Int64("size", fh.Size))
	return url, nil
}

// GeneratePresignedURL returns a time-limited S3 upload URL.
// This allows clients to upload directly to S3, bypassing your server.
func (s *S3Storage) GeneratePresignedURL(ctx context.Context, key string, contentType string) (string, error) {
	objectKey := s.cfg.BucketPath + key

	// Create a PutObject request
	req, _ := s.s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      &s.cfg.Bucket,
		Key:         &objectKey,
		ContentType: &contentType,
	})

	// Generate presigned URL (valid for configured duration)
	urlStr, err := req.Presign(time.Duration(s.cfg.PreSignedTTL) * time.Second)
	if err != nil {
		return "", fmt.Errorf("generating presigned URL: %w", err)
	}

	s.log.Debug("presigned URL generated", zap.String("key", objectKey))
	return urlStr, nil
}

// GetMaxFileSize returns the maximum allowed file size in bytes.
func (s *S3Storage) GetMaxFileSize() int64 {
	return s.cfg.MaxFileSize
}
