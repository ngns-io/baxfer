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
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ngns-io/baxfer/pkg/logger"
)

type S3Uploader struct {
	client *s3.Client
	bucket string
	log    logger.Logger
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
		client: client,
		bucket: bucket,
		log:    log,
	}

	// Log the provider initialization
	log.Info("Initialized storage provider",
		"provider", "AWS S3",
		"region", region,
		"bucket", bucket)

	return uploader, nil
}

func (u *S3Uploader) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	uploader := manager.NewUploader(u.client, func(u *manager.Uploader) {
		u.PartSize = 100 * 1024 * 1024
		u.Concurrency = 10
	})

	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:        &u.bucket,
		Key:           &key,
		Body:          reader,
		ContentLength: aws.Int64(size), // Convert int64 to *int64
	})
	return err
}

func (u *S3Uploader) Download(ctx context.Context, key string, writer io.Writer) error {
	output, err := u.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &u.bucket,
		Key:    &key,
	})
	if err != nil {
		return formatDownloadError("s3", key, err)
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

func (u *S3Uploader) List(ctx context.Context, prefix string) ([]string, error) {
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

func (u *S3Uploader) Delete(ctx context.Context, key string) error {
	_, err := u.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &u.bucket,
		Key:    &key,
	})
	return err
}

func (u *S3Uploader) FileExists(ctx context.Context, key string) (bool, error) {
	_, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &u.bucket,
		Key:    &key,
	})
	if err != nil {
		// Check for various forms of "not found" errors
		var (
			nsk      *types.NoSuchKey
			notFound *types.NotFound
		)
		if errors.As(err, &nsk) ||
			errors.As(err, &notFound) ||
			strings.Contains(err.Error(), "NotFound") ||
			strings.Contains(err.Error(), "status code: 404") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (u *S3Uploader) GetFileInfo(ctx context.Context, key string) (*FileInfo, error) {
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
