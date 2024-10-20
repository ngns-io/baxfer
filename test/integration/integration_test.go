package integration

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ngns-io/baxfer/pkg/logger"
	"github.com/ngns-io/baxfer/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLogger(t *testing.T) logger.Logger {
	log, err := logger.New(logger.LogConfig{
		Filename:     "integration_test.log",
		MaxSize:      10,
		MaxAge:       1,
		MaxBackups:   1,
		Compress:     false,
		ClearOnStart: true,
	}, false)
	require.NoError(t, err)
	return log
}

func TestIntegration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests")
	}

	log := setupLogger(t)
	defer log.Close()

	providers := []struct {
		name     string
		uploader func() (storage.Uploader, error)
	}{
		{"S3", func() (storage.Uploader, error) {
			return storage.NewS3Uploader(os.Getenv("AWS_REGION"), os.Getenv("AWS_BUCKET"))
		}},
		{"B2", func() (storage.Uploader, error) {
			return storage.NewB2Uploader(os.Getenv("B2_BUCKET"))
		}},
		{"R2", func() (storage.Uploader, error) {
			return storage.NewR2Uploader(os.Getenv("R2_BUCKET"))
		}},
	}

	for _, provider := range providers {
		t.Run(provider.name, func(t *testing.T) {
			uploader, err := provider.uploader()
			require.NoError(t, err)

			testData := []byte("test data " + time.Now().String())
			testKey := "test-file-" + provider.name + ".txt"
			compressedKey := strings.TrimSuffix(testKey, ".txt") + ".zip"

			// Test uncompressed upload
			err = uploader.Upload(context.Background(), testKey, bytes.NewReader(testData), int64(len(testData)))
			assert.NoError(t, err)

			// Test compressed upload
			err = uploader.Upload(context.Background(), compressedKey, bytes.NewReader(testData), int64(len(testData)))
			assert.NoError(t, err)

			// Test FileExists for both
			exists, err := uploader.FileExists(context.Background(), testKey)
			assert.NoError(t, err)
			assert.True(t, exists)

			exists, err = uploader.FileExists(context.Background(), compressedKey)
			assert.NoError(t, err)
			assert.True(t, exists)

			// Test GetFileInfo for both
			fileInfo, err := uploader.GetFileInfo(context.Background(), testKey)
			assert.NoError(t, err)
			assert.Equal(t, int64(len(testData)), fileInfo.Size)

			compressedInfo, err := uploader.GetFileInfo(context.Background(), compressedKey)
			assert.NoError(t, err)
			assert.True(t, compressedInfo.Size > 0)

			// Test Download for both
			var buf bytes.Buffer
			err = uploader.Download(context.Background(), testKey, &buf)
			assert.NoError(t, err)
			assert.Equal(t, testData, buf.Bytes())

			buf.Reset()
			err = uploader.Download(context.Background(), compressedKey, &buf)
			assert.NoError(t, err)
			assert.NotEmpty(t, buf.Bytes())

			// Test List
			files, err := uploader.List(context.Background(), "")
			assert.NoError(t, err)
			assert.Contains(t, files, testKey)
			assert.Contains(t, files, compressedKey)

			// Test Delete for both
			err = uploader.Delete(context.Background(), testKey)
			assert.NoError(t, err)

			err = uploader.Delete(context.Background(), compressedKey)
			assert.NoError(t, err)

			// Verify deletion for both
			exists, err = uploader.FileExists(context.Background(), testKey)
			assert.NoError(t, err)
			assert.False(t, exists)

			exists, err = uploader.FileExists(context.Background(), compressedKey)
			assert.NoError(t, err)
			assert.False(t, exists)
		})
	}
}
