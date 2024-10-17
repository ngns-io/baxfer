package storage

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/kurin/blazer/b2"
)

type B2Uploader struct {
	client *b2.Client
	bucket string
}

func NewB2Uploader(bucket string) (*B2Uploader, error) {
	ctx := context.Background()
	client, err := b2.NewClient(ctx, os.Getenv("B2_KEY_ID"), os.Getenv("B2_APP_KEY"))
	if err != nil {
		return nil, err
	}

	return &B2Uploader{
		client: client,
		bucket: bucket,
	}, nil
}

func (u *B2Uploader) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	b, err := u.client.Bucket(ctx, u.bucket)
	if err != nil {
		return err
	}

	obj := b.Object(key)
	w := obj.NewWriter(ctx)
	defer w.Close()

	_, err = io.Copy(w, reader)
	return err
}

func (u *B2Uploader) Download(ctx context.Context, key string, writer io.Writer) error {
	b, err := u.client.Bucket(ctx, u.bucket)
	if err != nil {
		return err
	}

	obj := b.Object(key)
	r := obj.NewReader(ctx)
	defer r.Close()

	_, err = io.Copy(writer, r)
	return err
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
		// Check if the error message indicates that the file doesn't exist
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return false, nil
		}
		// For any other error, return it
		return false, err
	}
	// If there's no error, the file exists
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
		LastModified: time.Unix(attrs.UploadTimestamp.Unix(), 0), // Convert to time.Time
		Size:         attrs.Size,
	}, nil
}
