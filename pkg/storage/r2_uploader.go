package storage

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/ngns-io/baxfer/pkg/logger"
)

type R2Uploader struct {
	session  *session.Session
	bucket   string
	s3Client *s3.S3
}

func NewR2Uploader(bucket string, log logger.Logger) (*R2Uploader, error) {
	accountID := os.Getenv("CF_ACCOUNT_ID")
	accessKeyID := os.Getenv("CF_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("CF_ACCESS_KEY_SECRET")

	if accountID == "" || accessKeyID == "" || accessKeySecret == "" {
		return nil, fmt.Errorf("Cloudflare R2 credentials not found in environment variables")
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyID, accessKeySecret, ""),
		Endpoint:    aws.String("https://" + accountID + ".r2.cloudflarestorage.com"),
		Region:      aws.String("auto"),
	})
	if err != nil {
		return nil, err
	}

	uploader := &R2Uploader{
		session:  sess,
		bucket:   bucket,
		s3Client: s3.New(sess),
	}

	// Log the provider initialization
	log.Info("Initialized storage provider",
		"provider", "Cloudflare R2",
		"account", accountID,
		"bucket", bucket)

	return uploader, nil
}

func (u *R2Uploader) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	uploader := s3manager.NewUploader(u.session, func(u *s3manager.Uploader) {
		u.PartSize = 100 * 1024 * 1024 // 100MB parts
		u.Concurrency = 10             // 10 concurrent uploads
	})

	_, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	return err
}

func (u *R2Uploader) Download(ctx context.Context, key string, writer io.Writer) error {
	downloader := s3manager.NewDownloader(u.session, func(d *s3manager.Downloader) {
		d.PartSize = 100 * 1024 * 1024 // 100MB parts
		d.Concurrency = 10             // 10 concurrent downloads
	})

	writerAt := newWriterAtWrapper(writer)
	_, err := downloader.DownloadWithContext(ctx, writerAt, &s3.GetObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (u *R2Uploader) List(ctx context.Context, prefix string) ([]string, error) {
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

func (u *R2Uploader) Delete(ctx context.Context, key string) error {
	_, err := u.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (u *R2Uploader) FileExists(ctx context.Context, key string) (bool, error) {
	_, err := u.s3Client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound", "NoSuchKey", "404":
				return false, nil
			default:
				return false, err
			}
		}
		return false, err
	}
	return true, nil
}

func (u *R2Uploader) GetFileInfo(ctx context.Context, key string) (*FileInfo, error) {
	result, err := u.s3Client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		LastModified: *result.LastModified,
		Size:         *result.ContentLength,
	}, nil
}
