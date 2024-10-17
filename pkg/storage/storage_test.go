package storage

import (
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
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "baxfer-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.bak")
	err = os.WriteFile(testFile, []byte("test data"), 0644)
	assert.NoError(t, err)

	mockUploader := new(MockUploader)
	mockLogger := NewMockLogger()

	mockUploader.On("FileExists", mock.Anything, "test.bak").Return(false, nil)
	mockUploader.On("Upload", mock.Anything, "test.bak", mock.Anything, int64(9)).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	app := &cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String("backupext", ".bak", "doc")
	set.Bool("non-interactive", true, "doc")
	ctx := cli.NewContext(app, set, nil)
	ctx.Set("backupext", ".bak")
	ctx.Set("non-interactive", "true")

	// Set the arguments for the context
	err = set.Parse([]string{tempDir})
	assert.NoError(t, err)

	err = Upload(ctx, mockUploader, mockLogger)
	assert.NoError(t, err)

	mockUploader.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileUploadEligible(t *testing.T) {
	mockUploader := new(MockUploader)
	mockLogger := NewMockLogger()

	// Test when file doesn't exist
	mockUploader.On("FileExists", mock.Anything, "test.bak").Return(false, nil).Once()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	eligible, err := fileUploadEligible(context.Background(), mockUploader, "test.bak", &mockFileInfo{modTime: time.Now(), size: 100}, mockLogger)
	assert.NoError(t, err)
	assert.True(t, eligible)

	// Test when file exists but local file is newer
	mockUploader.On("FileExists", mock.Anything, "test.bak").Return(true, nil).Once()
	localTime := time.Now()
	remoteTime := localTime.Add(-1 * time.Hour)
	mockUploader.On("GetFileInfo", mock.Anything, "test.bak").Return(&FileInfo{LastModified: remoteTime, Size: 100}, nil).Once()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	eligible, err = fileUploadEligible(context.Background(), mockUploader, "test.bak", &mockFileInfo{modTime: localTime, size: 100}, mockLogger)
	assert.NoError(t, err)
	assert.True(t, eligible)

	// Test when file exists, same modification time, but different size
	mockUploader.On("FileExists", mock.Anything, "test.bak").Return(true, nil).Once()
	mockUploader.On("GetFileInfo", mock.Anything, "test.bak").Return(&FileInfo{LastModified: localTime, Size: 200}, nil).Once()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	eligible, err = fileUploadEligible(context.Background(), mockUploader, "test.bak", &mockFileInfo{modTime: localTime, size: 100}, mockLogger)
	assert.NoError(t, err)
	assert.True(t, eligible)

	// Test when file exists, same modification time and size
	mockUploader.On("FileExists", mock.Anything, "test.bak").Return(true, nil).Once()
	mockUploader.On("GetFileInfo", mock.Anything, "test.bak").Return(&FileInfo{LastModified: localTime, Size: 100}, nil).Once()
	eligible, err = fileUploadEligible(context.Background(), mockUploader, "test.bak", &mockFileInfo{modTime: localTime, size: 100}, mockLogger)
	assert.NoError(t, err)
	assert.False(t, eligible)

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
