package storage

import (
	"context"
	"io"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Uploader struct {
	uploader *s3manager.Uploader
	bucket   string
	s3Client *s3.S3
}

func NewS3Uploader(region, bucket string) (*S3Uploader, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}

	return &S3Uploader{
		uploader: s3manager.NewUploader(sess),
		bucket:   bucket,
		s3Client: s3.New(sess),
	}, nil
}

func (u *S3Uploader) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	_, err := u.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	return err
}

// writerAtWrapper wraps an io.Writer and implements io.WriterAt
type writerAtWrapper struct {
	w  io.Writer
	mu sync.Mutex
}

func (waw *writerAtWrapper) WriteAt(p []byte, off int64) (n int, err error) {
	waw.mu.Lock()
	defer waw.mu.Unlock()
	return waw.w.Write(p)
}

func (u *S3Uploader) Download(ctx context.Context, key string, writer io.Writer) error {
	downloader := s3manager.NewDownloaderWithClient(u.s3Client)
	writerAt := &writerAtWrapper{w: writer}
	_, err := downloader.DownloadWithContext(ctx, writerAt, &s3.GetObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (u *S3Uploader) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	err := u.s3Client.ListObjectsV2PagesWithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(u.bucket),
		Prefix: aws.String(prefix),
	}, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			keys = append(keys, *obj.Key)
		}
		return true
	})
	return keys, err
}

func (u *S3Uploader) Delete(ctx context.Context, key string) error {
	_, err := u.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (u *S3Uploader) FileExists(ctx context.Context, key string) (bool, error) {
	_, err := u.s3Client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				return false, nil
			default:
				return false, err
			}
		}
		return false, err
	}
	return true, nil
}

func (u *S3Uploader) GetFileInfo(ctx context.Context, key string) (*FileInfo, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	}

	result, err := u.s3Client.HeadObjectWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		LastModified: *result.LastModified,
		Size:         *result.ContentLength,
	}, nil
}
