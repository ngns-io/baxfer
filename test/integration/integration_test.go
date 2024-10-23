package integration

import (
	"bytes"
	"context"
	"fmt"
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
		name            string
		uploader        func() (storage.Uploader, error)
		requiredEnvVars []string
	}{
		{
			name: "S3",
			uploader: func() (storage.Uploader, error) {
				return storage.NewS3Uploader(os.Getenv("AWS_REGION"), os.Getenv("AWS_BUCKET"), log)
			},
			requiredEnvVars: []string{"AWS_REGION", "AWS_BUCKET", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"},
		},
		{
			name: "B2",
			uploader: func() (storage.Uploader, error) {
				return storage.NewB2Uploader(os.Getenv("B2_BUCKET"), log)
			},
			requiredEnvVars: []string{"B2_BUCKET", "B2_KEY_ID", "B2_APP_KEY"},
		},
		{
			name: "B2S3",
			uploader: func() (storage.Uploader, error) {
				return storage.NewB2S3Uploader(os.Getenv("AWS_REGION"), os.Getenv("B2_BUCKET"), log)
			},
			requiredEnvVars: []string{"AWS_REGION", "B2_BUCKET", "B2_KEY_ID", "B2_APP_KEY"},
		},
		{
			name: "R2",
			uploader: func() (storage.Uploader, error) {
				return storage.NewR2Uploader(os.Getenv("R2_BUCKET"), log)
			},
			requiredEnvVars: []string{"R2_BUCKET", "CF_ACCOUNT_ID", "CF_ACCESS_KEY_ID", "CF_ACCESS_KEY_SECRET"},
		},
		{
			name: "SFTP",
			uploader: func() (storage.Uploader, error) {
				port := 22
				if portStr := os.Getenv("SFTP_PORT"); portStr != "" {
					fmt.Sscanf(portStr, "%d", &port)
				}
				return storage.NewSFTPUploader(
					os.Getenv("SFTP_HOST"),
					port,
					os.Getenv("SFTP_USER"),
					os.Getenv("SFTP_PATH"),
					log,
				)
			},
			requiredEnvVars: []string{"SFTP_HOST", "SFTP_USER", "SFTP_PATH"},
		},
	}

	for _, provider := range providers {
		t.Run(provider.name, func(t *testing.T) {
			// Skip if required environment variables are not set
			for _, env := range provider.requiredEnvVars {
				if os.Getenv(env) == "" {
					t.Skipf("Skipping %s tests: %s not set", provider.name, env)
				}
			}

			// Additional check for SFTP authentication
			if provider.name == "SFTP" {
				if os.Getenv("SFTP_PRIVATE_KEY") == "" && os.Getenv("SFTP_PASSWORD") == "" {
					t.Skip("Skipping SFTP tests: neither SFTP_PRIVATE_KEY nor SFTP_PASSWORD is set")
				}
			}

			uploader, err := provider.uploader()
			require.NoError(t, err)

			// Close the SFTP connection after tests if applicable
			if closer, ok := uploader.(interface{ Close() error }); ok {
				defer closer.Close()
			}

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
