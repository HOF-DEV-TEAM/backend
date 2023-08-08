package uploader

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"bitbucket.org/hofng/hofApp/infrastructure/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"
)

type AWSClient struct {
	Config     *config.ServerConfig
	Log        *zap.Logger
	Session    *session.Session
	S3Uploader *s3manager.Uploader
}

func (awsClient *AWSClient) ConnectAWS() {
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}
	awsConfig := &aws.Config{
		Region: aws.String(awsClient.Config.AwsConfiguration.Region),
		Credentials: credentials.NewStaticCredentials(
			awsClient.Config.AwsConfiguration.Endpoint,
			awsClient.Config.AwsConfiguration.Secret,
			"",
		),
		HTTPClient: httpClient,
	}
	awsSession := session.New(awsConfig)

	awsClient.Session = awsSession
	awsClient.S3Uploader = s3manager.NewUploader(awsSession)
}

func (awsClient *AWSClient) Upload(ctx context.Context, fileHandler *FileHandler, bucketKey string) (*s3manager.UploadOutput, error) {
	input := &s3manager.UploadInput{
		Bucket:      aws.String(string(awsClient.Config.AwsConfiguration.Bucket)),     // bucket's name
		Key:         aws.String(fmt.Sprintf("%s%s", bucketKey, fileHandler.FileName)), // files destination location
		Body:        bytes.NewReader(fileHandler.File),                                // content of the file
		ContentType: aws.String(fileHandler.ContentType),
		//ACL:         aws.String("public-read"), // content type
	}
	return awsClient.S3Uploader.UploadWithContext(ctx, input)
}
