// This file implements the local filesystem storage backend for uploaded
// files.

package file

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/grand"

	"lina-core/pkg/closeutil"
)

// LocalStorage implements Storage interface using local file system.
type LocalStorage struct {
	basePath string // Base directory for file storage, e.g. "temp/upload"
}

// NewLocalStorage creates a LocalStorage instance.
// The caller should pass the upload path resolved during service construction.
func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

// Put saves file data organized by year/month directory structure.
// Returns the relative path from basePath, e.g. "2026/03/20260319_abc12345.png".
func (s *LocalStorage) Put(ctx context.Context, filename string, data io.Reader) (path string, err error) {
	now := gtime.Now()
	dir := fmt.Sprintf("%s/%s", now.Format("Y"), now.Format("m"))
	fullDir := gfile.Join(s.basePath, dir)
	if err := gfile.Mkdir(fullDir); err != nil {
		return "", err
	}

	ext := gfile.ExtName(filename)
	storedName := fmt.Sprintf("%s_%s", now.Format("Ymd_His"), grand.S(8))
	if ext != "" {
		storedName = storedName + "." + gstr.ToLower(ext)
	}

	fullPath := gfile.Join(fullDir, storedName)
	f, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer closeutil.Close(f, &err, "关闭上传目标文件失败")

	if _, err = io.Copy(f, data); err != nil {
		if removeErr := os.Remove(fullPath); removeErr != nil && !os.IsNotExist(removeErr) {
			return "", gerror.Wrapf(err, "写入文件失败，且清理临时文件失败: %v", removeErr)
		}
		return "", err
	}

	relativePath := gfile.Join(dir, storedName)
	return relativePath, nil
}

// Get opens the file at the given relative path for reading.
func (s *LocalStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := gfile.Join(s.basePath, path)
	return os.Open(fullPath)
}

// Delete removes the file at the given relative path.
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := gfile.Join(s.basePath, path)
	if !gfile.Exists(fullPath) {
		return nil
	}
	return os.Remove(fullPath)
}

// Url returns the public access URL for the file.
func (s *LocalStorage) Url(ctx context.Context, path string) string {
	return "/api/v1/uploads/" + path
}
