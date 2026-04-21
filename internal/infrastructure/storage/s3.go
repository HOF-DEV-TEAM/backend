package storage

import (
	"context"
	"fmt"
	"mime/multipart"

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
	cfg      config.AWSConfig
	log      *zap.Logger
}

// NewS3Storage connects to AWS and returns a ready-to-use S3Storage.
func NewS3Storage(cfg config.AWSConfig, log *zap.Logger) (*S3Storage, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cfg.Region),
		Credentials: credentials.NewStaticCredentials(cfg.AccessKey, cfg.Secret, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("creating AWS session: %w", err)
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
		return "", fmt.Errorf("opening upload file: %w", err)
	}
	defer f.Close()

	objectKey := s.cfg.BucketPath + key

	_, err = s.uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.cfg.Bucket),
		Key:         aws.String(objectKey),
		Body:        f,
		ContentType: aws.String(fh.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", fmt.Errorf("uploading to S3: %w", err)
	}

	url := fmt.Sprintf("%s%s/%s", s.cfg.BaseURL, s.cfg.Bucket, objectKey)
	s.log.Info("file uploaded to S3", zap.String("url", url))
	return url, nil
}
