package storage

import (
	"archive/zip"
	"bytes"
	"context"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ngns-io/baxfer/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/urfave/cli/v2"
)

// MockLogger is a mock implementation of the logger.Logger
type MockLogger struct {
	mock.Mock
	logger.Logger
}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Fatal(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockUploader is a mock implementation of the Uploader interface
type MockUploader struct {
	mock.Mock
}

func (m *MockUploader) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	args := m.Called(ctx, key, reader, size)
	return args.Error(0)
}

func (m *MockUploader) Download(ctx context.Context, key string, writer io.Writer) error {
	args := m.Called(ctx, key, writer)
	return args.Error(0)
}

func (m *MockUploader) List(ctx context.Context, prefix string) ([]string, error) {
	args := m.Called(ctx, prefix)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockUploader) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockUploader) FileExists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockUploader) GetFileInfo(ctx context.Context, key string) (*FileInfo, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(*FileInfo), args.Error(1)
}

func TestUpload(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "baxfer-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.bak")
	err = os.WriteFile(testFile, []byte("test data"), 0644)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		compress    bool
		expectedKey string
	}{
		{"Uncompressed", false, "test.bak"},
		{"Compressed", true, "test.zip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUploader := new(MockUploader)
			mockLogger := NewMockLogger()

			uploadedData := &bytes.Buffer{}

			mockUploader.On("FileExists", mock.Anything, tt.expectedKey).Return(false, nil)
			mockUploader.On("Upload",
				mock.Anything,
				tt.expectedKey,
				mock.MatchedBy(func(r interface{}) bool {
					reader, ok := r.(io.Reader)
					if !ok {
						return false
					}
					// Synchronously read all data
					_, err := io.Copy(uploadedData, reader)
					return err == nil
				}),
				mock.AnythingOfType("int64"),
			).Return(nil)

			mockLogger.On("Info", mock.Anything, mock.Anything).Return()

			app := &cli.App{}
			set := flag.NewFlagSet("test", 0)
			set.String("backupext", ".bak", "doc")
			set.Bool("compress", tt.compress, "doc")
			set.Bool("non-interactive", true, "doc")
			ctx := cli.NewContext(app, set, nil)

			err := set.Parse([]string{tempDir})
			assert.NoError(t, err)

			err = Upload(ctx, mockUploader, mockLogger)
			assert.NoError(t, err)

			if tt.compress {
				// Verify uploaded data is a valid zip file
				reader := bytes.NewReader(uploadedData.Bytes())
				zipReader, err := zip.NewReader(reader, int64(uploadedData.Len()))
				assert.NoError(t, err)
				assert.Len(t, zipReader.File, 1)
			} else {
				assert.Equal(t, "test data", uploadedData.String())
			}

			mockUploader.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestFileUploadEligible(t *testing.T) {
	mockUploader := new(MockUploader)
	mockLogger := NewMockLogger()

	tests := []struct {
		name           string
		key            string
		fileExists     bool
		localModTime   time.Time
		remoteModTime  time.Time
		localSize      int64
		remoteSize     int64
		expectedResult bool
		expectInfoLog  bool
	}{
		{
			name:           "Uncompressed file doesn't exist remotely",
			key:            "test.bak",
			fileExists:     false,
			expectedResult: true,
			expectInfoLog:  false,
		},
		{
			name:           "Compressed file doesn't exist remotely",
			key:            "test.zip",
			fileExists:     false,
			expectedResult: true,
			expectInfoLog:  false,
		},
		{
			name:           "Uncompressed file exists, local is newer",
			key:            "test.bak",
			fileExists:     true,
			localModTime:   time.Now(),
			remoteModTime:  time.Now().Add(-1 * time.Hour),
			localSize:      100,
			remoteSize:     100,
			expectedResult: true,
			expectInfoLog:  true,
		},
		{
			name:           "Compressed file exists, same mod time, different size",
			key:            "test.zip",
			fileExists:     true,
			localModTime:   time.Now(),
			remoteModTime:  time.Now(),
			localSize:      100,
			remoteSize:     200,
			expectedResult: true,
			expectInfoLog:  true,
		},
		{
			name:           "Uncompressed file exists, same mod time and size",
			key:            "test.bak",
			fileExists:     true,
			localModTime:   time.Now(),
			remoteModTime:  time.Now(),
			localSize:      100,
			remoteSize:     100,
			expectedResult: false,
			expectInfoLog:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUploader.On("FileExists", mock.Anything, tt.key).Return(tt.fileExists, nil).Once()
			if tt.fileExists {
				mockUploader.On("GetFileInfo", mock.Anything, tt.key).Return(&FileInfo{LastModified: tt.remoteModTime, Size: tt.remoteSize}, nil).Once()
			}
			if tt.expectInfoLog {
				mockLogger.On("Info", mock.Anything, mock.Anything).Return()
			}

			eligible, err := fileUploadEligible(context.Background(), mockUploader, tt.key, &mockFileInfo{modTime: tt.localModTime, size: tt.localSize}, mockLogger)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, eligible)

			mockUploader.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestDownload(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "baxfer-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	mockUploader := new(MockUploader)
	mockLogger := NewMockLogger()

	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String("output", filepath.Join(tempDir, "downloaded.bak"), "doc")
	ctx := cli.NewContext(app, set, nil)
	ctx.Set("output", filepath.Join(tempDir, "downloaded.bak"))

	mockUploader.On("Download", mock.Anything, "test.bak", mock.Anything).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	err = set.Parse([]string{"test.bak"})
	assert.NoError(t, err)

	err = Download(ctx, mockUploader, mockLogger)
	assert.NoError(t, err)

	mockUploader.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestPrune(t *testing.T) {
	mockUploader := new(MockUploader)
	mockLogger := NewMockLogger()

	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String("age", "24h", "doc")
	ctx := cli.NewContext(app, set, nil)
	ctx.Set("age", "24h")

	oldFiles := []string{"old1.bak", "old2.bak"}
	mockUploader.On("List", mock.Anything, "").Return(oldFiles, nil)
	mockUploader.On("GetFileInfo", mock.Anything, mock.Anything).Return(&FileInfo{LastModified: time.Now().Add(-48 * time.Hour)}, nil)
	mockUploader.On("Delete", mock.Anything, mock.Anything).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	err := Prune(ctx, mockUploader, mockLogger)
	assert.NoError(t, err)

	mockUploader.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// mockFileInfo is a mock implementation of os.FileInfo for testing
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
	sys     interface{}
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return m.sys }
