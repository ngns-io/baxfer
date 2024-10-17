// integration_test.go

package main

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/ngns-io/baxfer/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestS3Integration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests")
	}

	uploader, err := storage.NewS3Uploader("us-east-1", "test-bucket")
	assert.NoError(t, err)

	// Create a test file
	testData := []byte("test data")
	testKey := "test-file.txt"

	// Test Upload
	err = uploader.Upload(context.Background(), testKey, bytes.NewReader(testData), int64(len(testData)))
	assert.NoError(t, err)

	// Test FileExists
	exists, err := uploader.FileExists(context.Background(), testKey)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test Download
	var buf bytes.Buffer
	err = uploader.Download(context.Background(), testKey, &buf)
	assert.NoError(t, err)
	assert.Equal(t, testData, buf.Bytes())

	// Test List
	files, err := uploader.List(context.Background(), "")
	assert.NoError(t, err)
	assert.Contains(t, files, testKey)

	// Test Delete
	err = uploader.Delete(context.Background(), testKey)
	assert.NoError(t, err)

	// Verify deletion
	exists, err = uploader.FileExists(context.Background(), testKey)
	assert.NoError(t, err)
	assert.False(t, exists)
}
