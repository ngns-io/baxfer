package storage

import (
	"io"
	"sync"
	"time"
)

// writerAtWrapper adapts an io.Writer to satisfy the io.WriterAt interface.
// Used for concurrent downloads in cloud storage providers.
type writerAtWrapper struct {
	w  io.Writer
	mu sync.Mutex
}

func newWriterAtWrapper(w io.Writer) *writerAtWrapper {
	return &writerAtWrapper{w: w}
}

func (waw *writerAtWrapper) WriteAt(p []byte, off int64) (n int, err error) {
	waw.mu.Lock()
	defer waw.mu.Unlock()
	return waw.w.Write(p)
}

// FileInfo represents metadata about a file in cloud storage
type FileInfo struct {
	LastModified time.Time
	Size         int64
}
