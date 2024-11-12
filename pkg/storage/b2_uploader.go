package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Backblaze/blazer/b2"
	"github.com/ngns-io/baxfer/pkg/logger"
)

type B2Uploader struct {
	client *b2.Client
	bucket string
}

func NewB2Uploader(bucket string, log logger.Logger) (*B2Uploader, error) {
	ctx := context.Background()
	client, err := b2.NewClient(ctx, os.Getenv("B2_KEY_ID"), os.Getenv("B2_APP_KEY"))
	if err != nil {
		return nil, err
	}

	uploader := &B2Uploader{
		client: client,
		bucket: bucket,
	}

	// Log the provider initialization
	log.Info("Initialized storage provider",
		"provider", "Backblaze B2",
		"bucket", bucket)

	return uploader, nil
}

func (u *B2Uploader) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	b, err := u.client.Bucket(ctx, u.bucket)
	if err != nil {
		return err
	}

	w := b.Object(key).NewWriter(ctx)
	w.ConcurrentUploads = 5 // Number of concurrent upload threads
	defer w.Close()

	_, err = io.Copy(w, reader)
	return err
}

func (u *B2Uploader) Download(ctx context.Context, key string, writer io.Writer) error {
	b, err := u.client.Bucket(ctx, u.bucket)
	if err != nil {
		return formatDownloadError("b2", key, err)
	}

	r := b.Object(key).NewReader(ctx)
	defer r.Close()

	// B2 reader automatically handles concurrent downloads
	_, err = io.Copy(writer, r)
	if err != nil {
		return &UserError{
			Message: fmt.Sprintf("Error reading file content: %s", key),
			Cause:   err,
		}
	}
	return nil
}

func (u *B2Uploader) List(ctx context.Context, prefix string) ([]string, error) {
	b, err := u.client.Bucket(ctx, u.bucket)
	if err != nil {
		return nil, err
	}

	var keys []string
	iter := b.List(ctx, b2.ListPrefix(prefix))
	for iter.Next() {
		keys = append(keys, iter.Object().Name())
	}
	return keys, iter.Err()
}

func (u *B2Uploader) Delete(ctx context.Context, key string) error {
	b, err := u.client.Bucket(ctx, u.bucket)
	if err != nil {
		return err
	}

	obj := b.Object(key)
	return obj.Delete(ctx)
}

func (u *B2Uploader) FileExists(ctx context.Context, key string) (bool, error) {
	b, err := u.client.Bucket(ctx, u.bucket)
	if err != nil {
		return false, err
	}

	obj := b.Object(key)
	_, err = obj.Attrs(ctx)
	if err != nil {
		// Check for 404 status in error message
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (u *B2Uploader) GetFileInfo(ctx context.Context, key string) (*FileInfo, error) {
	b, err := u.client.Bucket(ctx, u.bucket)
	if err != nil {
		return nil, err
	}

	obj := b.Object(key)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		LastModified: attrs.UploadTimestamp,
		Size:         attrs.Size,
	}, nil
}
