package file

import (
	"context"
	"io"
)

// Storage defines the interface for file storage backends.
// Default implementation is local file system storage.
// Can be extended to support OSS (Aliyun, Tencent, MinIO, etc.).
type Storage interface {
	// Put saves file data and returns the relative storage path.
	Put(ctx context.Context, filename string, data io.Reader) (path string, err error)

	// Get reads file data from the given storage path.
	Get(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes the file at the given storage path.
	Delete(ctx context.Context, path string) error

	// Url returns the public access URL for the given storage path.
	Url(ctx context.Context, path string) string
}
