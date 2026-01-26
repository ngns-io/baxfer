package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ngns-io/baxfer/pkg/logger"
)

type R2Uploader struct {
	*S3CompatibleUploader
}

func NewR2Uploader(bucket string, log logger.Logger) (*R2Uploader, error) {
	accountID := os.Getenv("CF_ACCOUNT_ID")
	accessKeyID := os.Getenv("CF_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("CF_ACCESS_KEY_SECRET")

	if accountID == "" || accessKeyID == "" || accessKeySecret == "" {
		return nil, fmt.Errorf("Cloudflare R2 credentials not found in environment variables")
	}

	customEndpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID, accessKeySecret, "",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(customEndpoint)
		o.UsePathStyle = true
	})

	uploader := &R2Uploader{
		S3CompatibleUploader: NewS3CompatibleUploader(
			client,
			bucket,
			"r2",
			log,
			100*1024*1024, // 100MB part size
			5,             // concurrency
		),
	}

	log.Info("Initialized storage provider",
		"provider", "Cloudflare R2",
		"account", accountID,
		"bucket", bucket)

	return uploader, nil
}

func (u *R2Uploader) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	input := &s3.PutObjectInput{
		Bucket:        &u.Bucket,
		Key:           &key,
		Body:          reader,
		ContentLength: aws.Int64(size),
	}

	// Force path-style addressing for R2
	_, err := u.Client.PutObject(ctx, input, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	return err
}

// FileExists overrides the base implementation with R2-specific error handling
func (u *R2Uploader) FileExists(ctx context.Context, key string) (bool, error) {
	_, err := u.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &u.Bucket,
		Key:    &key,
	})
	if err != nil {
		u.Log.Debug("HeadObject error details",
			"error", err,
			"bucket", u.Bucket,
			"key", key)

		var (
			nsk      *types.NoSuchKey
			notFound *types.NotFound
		)
		if errors.As(err, &nsk) ||
			errors.As(err, &notFound) ||
			strings.Contains(err.Error(), "NotFound") ||
			strings.Contains(err.Error(), "status code: 404") ||
			strings.Contains(err.Error(), "StatusCode: 404") {
			return false, nil
		}

		// For R2-specific errors, check the status code
		if strings.Contains(err.Error(), "StatusCode: 411") ||
			strings.Contains(err.Error(), "MissingContentLength") {
			u.Log.Warn("Unexpected 411 error from R2 HeadObject",
				"bucket", u.Bucket,
				"key", key)
			return false, nil
		}

		return false, fmt.Errorf("error checking if file exists: %w", err)
	}
	return true, nil
}
