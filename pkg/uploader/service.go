package uploader

import (
	"bitbucket.org/hofng/hofApp/infrastructure/config"
	"context"
	"go.uber.org/zap"
)

type FileHandler struct {
	FileName    string
	FileMIME    string
	FileSize    int64
	ContentType string
	File        []byte
}

type Service interface {
	UploadFile(ctx context.Context, fileHandler *FileHandler, bucketKey string) (string, error)
}

type uploadService struct {
	// repo   Repository
	awsClient *AWSClient
	log       *zap.Logger
	config    *config.ServerConfig
}

func NewService(awsClient *AWSClient) Service {
	return &uploadService{awsClient: awsClient}
}

func (uploadSvc *uploadService) UploadFile(ctx context.Context, fileHandler *FileHandler, bucketKey string) (string, error) {
	output, err := uploadSvc.awsClient.Upload(ctx, fileHandler, bucketKey)
	if err != nil {
		return "", err
	}

	return output.Location, nil
}
