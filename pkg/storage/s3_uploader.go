package storage

import (
	"context"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ngns-io/baxfer/pkg/logger"
)

type S3Uploader struct {
	*S3CompatibleUploader
}

func NewS3Uploader(region, bucket string, log logger.Logger) (*S3Uploader, error) {
	// Use provided region, or AWS_REGION env var, or default to us-east-1
	if region == "" {
		region = os.Getenv("AWS_REGION")
		if region == "" {
			region = "us-east-1"
		}
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)

	uploader := &S3Uploader{
		S3CompatibleUploader: NewS3CompatibleUploader(
			client,
			bucket,
			"s3",
			log,
			100*1024*1024, // 100MB part size
			10,            // concurrency
		),
	}

	log.Info("Initialized storage provider",
		"provider", "AWS S3",
		"region", region,
		"bucket", bucket)

	return uploader, nil
}

func (u *S3Uploader) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	_, err := u.Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:        &u.Bucket,
		Key:           &key,
		Body:          reader,
		ContentLength: aws.Int64(size),
	})
	return err
}
