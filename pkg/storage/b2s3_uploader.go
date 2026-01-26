package storage

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ngns-io/baxfer/pkg/logger"
)

type B2S3Uploader struct {
	*S3CompatibleUploader
}

func NewB2S3Uploader(region, bucket string, log logger.Logger) (*B2S3Uploader, error) {
	keyID := os.Getenv("B2_KEY_ID")
	appKey := os.Getenv("B2_APP_KEY")

	if keyID == "" || appKey == "" {
		return nil, fmt.Errorf("Backblaze B2 credentials not found in environment variables")
	}

	// Use provided region, or AWS_REGION env var, or default to us-west-002
	if region == "" {
		region = os.Getenv("AWS_REGION")
		if region == "" {
			region = "us-west-002"
		}
	}

	customEndpoint := fmt.Sprintf("https://s3.%s.backblazeb2.com", region)

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			keyID, appKey, "",
		)),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(customEndpoint)
		o.UsePathStyle = true
	})

	uploader := &B2S3Uploader{
		S3CompatibleUploader: NewS3CompatibleUploader(
			client,
			bucket,
			"b2s3",
			log,
			100*1024*1024, // 100MB part size
			5,             // concurrency
		),
	}

	log.Info("Initialized storage provider",
		"provider", "Backblaze B2 S3",
		"region", region,
		"bucket", bucket)

	return uploader, nil
}

func (u *B2S3Uploader) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	input := &s3.PutObjectInput{
		Bucket:        &u.Bucket,
		Key:           &key,
		Body:          reader,
		ContentLength: aws.Int64(size),
	}

	_, err := u.Uploader.Upload(ctx, input)
	return err
}
