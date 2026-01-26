package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ngns-io/baxfer/pkg/logger"
)

// S3CompatibleUploader contains shared implementation for S3-compatible storage providers
type S3CompatibleUploader struct {
	Client       *s3.Client
	Bucket       string
	Log          logger.Logger
	ProviderName string
	Uploader     *manager.Uploader
}

// NewS3CompatibleUploader creates a new S3CompatibleUploader with the given settings
func NewS3CompatibleUploader(client *s3.Client, bucket, providerName string, log logger.Logger, partSize int64, concurrency int) *S3CompatibleUploader {
	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.PartSize = partSize
		u.Concurrency = concurrency
	})

	return &S3CompatibleUploader{
		Client:       client,
		Bucket:       bucket,
		Log:          log,
		ProviderName: providerName,
		Uploader:     uploader,
	}
}

func (u *S3CompatibleUploader) Download(ctx context.Context, key string, writer io.Writer) error {
	output, err := u.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &u.Bucket,
		Key:    &key,
	})
	if err != nil {
		return formatDownloadError(u.ProviderName, key, err)
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

func (u *S3CompatibleUploader) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	paginator := s3.NewListObjectsV2Paginator(u.Client, &s3.ListObjectsV2Input{
		Bucket: &u.Bucket,
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

func (u *S3CompatibleUploader) Delete(ctx context.Context, key string) error {
	_, err := u.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &u.Bucket,
		Key:    &key,
	})
	return err
}

func (u *S3CompatibleUploader) FileExists(ctx context.Context, key string) (bool, error) {
	_, err := u.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &u.Bucket,
		Key:    &key,
	})
	if err != nil {
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
		return false, err
	}
	return true, nil
}

func (u *S3CompatibleUploader) GetFileInfo(ctx context.Context, key string) (*FileInfo, error) {
	output, err := u.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &u.Bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}

	if output.LastModified == nil || output.ContentLength == nil {
		return nil, fmt.Errorf("incomplete response from %s HeadObject for key: %s", u.ProviderName, key)
	}

	return &FileInfo{
		LastModified: *output.LastModified,
		Size:         *output.ContentLength,
	}, nil
}
