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
	client *s3.Client
	bucket string
	log    logger.Logger
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
		client: client,
		bucket: bucket,
		log:    log,
	}

	log.Info("Initialized storage provider",
		"provider", "Cloudflare R2",
		"account", accountID,
		"bucket", bucket)

	return uploader, nil
}

func (u *R2Uploader) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	input := &s3.PutObjectInput{
		Bucket:        &u.bucket,
		Key:           &key,
		Body:          reader,
		ContentLength: aws.Int64(size),
	}

	// Force path-style addressing for R2
	_, err := u.client.PutObject(ctx, input, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	return err
}

func (u *R2Uploader) Download(ctx context.Context, key string, writer io.Writer) error {
	output, err := u.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &u.bucket,
		Key:    &key,
	})
	if err != nil {
		return formatDownloadError("r2", key, err)
	}
	defer output.Body.Close()

	_, err = io.Copy(writer, output.Body)
	if err != nil {
		return &UserError{
			Message: fmt.Sprintf("Error reading file content: %s", key),
			Cause:   err,
		}
	}
	return nil
}

func (u *R2Uploader) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	paginator := s3.NewListObjectsV2Paginator(u.client, &s3.ListObjectsV2Input{
		Bucket: &u.bucket,
		Prefix: &prefix,
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			keys = append(keys, *obj.Key)
		}
	}

	return keys, nil
}

func (u *R2Uploader) Delete(ctx context.Context, key string) error {
	_, err := u.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &u.bucket,
		Key:    &key,
	})
	return err
}

func (u *R2Uploader) FileExists(ctx context.Context, key string) (bool, error) {
	_, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &u.bucket,
		Key:    &key,
	})
	if err != nil {
		// Log the raw error for debugging
		u.log.Debug("HeadObject error details",
			"error", err,
			"bucket", u.bucket,
			"key", key)

		// Check for any type of "not found" response
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
			// Log this unexpected condition
			u.log.Warn("Unexpected 411 error from R2 HeadObject",
				"bucket", u.bucket,
				"key", key)
			return false, nil // Assume file doesn't exist
		}

		return false, fmt.Errorf("error checking if file exists: %w", err)
	}
	return true, nil
}

func (u *R2Uploader) GetFileInfo(ctx context.Context, key string) (*FileInfo, error) {
	output, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &u.bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		LastModified: *output.LastModified,
		Size:         *output.ContentLength,
	}, nil
}
