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
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ngns-io/baxfer/pkg/logger"
)

type B2S3Uploader struct {
	client *s3.Client
	bucket string
	log    logger.Logger
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
		client: client,
		bucket: bucket,
		log:    log,
	}

	// Log the provider initialization
	log.Info("Initialized storage provider",
		"provider", "Backblaze B2 S3",
		"region", region,
		"bucket", bucket)

	return uploader, nil
}

func (u *B2S3Uploader) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	uploader := manager.NewUploader(u.client, func(u *manager.Uploader) {
		u.PartSize = 100 * 1024 * 1024 // 100MB part size
		u.Concurrency = 5              // Number of concurrent uploads
	})

	input := &s3.PutObjectInput{
		Bucket:        &u.bucket,
		Key:           &key,
		Body:          reader,
		ContentLength: aws.Int64(size),
	}

	_, err := uploader.Upload(ctx, input, func(u *manager.Uploader) {
		u.ClientOptions = append(u.ClientOptions, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	})

	return err
}

func (u *B2S3Uploader) Download(ctx context.Context, key string, writer io.Writer) error {
	output, err := u.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &u.bucket,
		Key:    &key,
	})
	if err != nil {
		return err
	}
	defer output.Body.Close()

	_, err = io.Copy(writer, output.Body)
	return err
}

func (u *B2S3Uploader) List(ctx context.Context, prefix string) ([]string, error) {
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

func (u *B2S3Uploader) Delete(ctx context.Context, key string) error {
	_, err := u.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &u.bucket,
		Key:    &key,
	})
	return err
}

func (u *B2S3Uploader) FileExists(ctx context.Context, key string) (bool, error) {
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

func (u *B2S3Uploader) GetFileInfo(ctx context.Context, key string) (*FileInfo, error) {
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
