package storage

import (
	"time"
)

// FileInfo represents metadata about a file in cloud storage
type FileInfo struct {
	LastModified time.Time
	Size         int64
}
