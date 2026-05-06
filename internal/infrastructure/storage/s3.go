// Package storage provides AWS S3 upload helpers.
package storage

import (
	"context"
	"fmt"
	"mime/multipart"

	"bitbucket.org/hofng/hofApp/internal/domain/shared"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"
)

// S3Storage uploads files to AWS S3.
type S3Storage struct {
	uploader *s3manager.Uploader
	cfg      *config.AWSConfig
	log      *zap.Logger
}

// Ensure S3Storage implements the Storage interface
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

	return &S3Storage{
		uploader: s3manager.NewUploader(sess),
		cfg:      cfg,
		log:      log,
	}, nil
}

// Upload stores the file in S3 under <bucketPath><key> and returns the public URL.
func (s *S3Storage) Upload(_ context.Context, fh *multipart.FileHeader, key string) (string, error) {
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

	_, err = s.uploader.Upload(&s3manager.UploadInput{
		Bucket:      &s.cfg.Bucket,
		Key:         &objectKey,
		Body:        f,
		ContentType: new(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("s3 upload: %w", err)
	}

	url := fmt.Sprintf("%s%s/%s", s.cfg.BaseURL, s.cfg.Bucket, objectKey)
	s.log.Info("file uploaded to S3", zap.String("url", url))
	return url, nil
}
