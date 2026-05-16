// This file contains safe file-open helpers for download and URL access,
// including storage-path normalization and content-type projection.

package file

import (
	"context"
	"errors"
	"io/fs"
	"mime"
	"path"
	"strings"

	"github.com/gogf/gf/v2/text/gstr"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
)

// OpenByID opens a stored file stream by metadata ID for download.
func (s *serviceImpl) OpenByID(ctx context.Context, id int64) (*OpenOutput, error) {
	fileInfo, err := s.Info(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.openStoredFile(ctx, fileInfo)
}

// OpenByPath opens a stored file stream by metadata storage path for URL access.
func (s *serviceImpl) OpenByPath(ctx context.Context, storagePath string) (*OpenOutput, error) {
	normalizedPath := normalizeStoragePath(storagePath)
	if normalizedPath == "" {
		return nil, bizerr.NewCode(CodeFileNotFound)
	}
	var fileInfo *entity.SysFile
	model := dao.SysFile.Ctx(ctx).Where(dao.SysFile.Columns().Path, normalizedPath)
	model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
	err := model.Scan(&fileInfo)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeFileRecordQueryFailed)
	}
	if fileInfo == nil {
		return nil, bizerr.NewCode(CodeFileNotFound)
	}
	return s.openStoredFile(ctx, fileInfo)
}

// openStoredFile opens the object represented by file metadata through the
// configured storage backend and attaches response metadata.
func (s *serviceImpl) openStoredFile(ctx context.Context, fileInfo *entity.SysFile) (*OpenOutput, error) {
	if fileInfo == nil || strings.TrimSpace(fileInfo.Path) == "" {
		return nil, bizerr.NewCode(CodeFileNotFound)
	}
	reader, err := s.storage.Get(ctx, fileInfo.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, bizerr.NewCode(CodeFileNotFound)
		}
		return nil, bizerr.WrapCode(err, CodeFileStorageReadFailed)
	}
	return &OpenOutput{
		Reader:      reader,
		Original:    fileInfo.Original,
		Suffix:      fileInfo.Suffix,
		ContentType: contentTypeForSuffix(fileInfo.Suffix),
		Size:        fileInfo.Size,
	}, nil
}

// normalizeStoragePath converts a URL path segment into a relative object key
// and rejects absolute or parent-directory paths before any storage access.
func normalizeStoragePath(raw string) string {
	candidate := strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
	candidate = strings.TrimPrefix(candidate, "/")
	if candidate == "" {
		return ""
	}
	cleaned := path.Clean(candidate)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return ""
	}
	return cleaned
}

// contentTypeForSuffix returns a safe content type for browser access and
// download responses based on stored file metadata.
func contentTypeForSuffix(suffix string) string {
	normalizedSuffix := strings.TrimPrefix(gstr.ToLower(strings.TrimSpace(suffix)), ".")
	switch normalizedSuffix {
	case "jpg", "jpeg", "png", "gif", "webp", "svg", "pdf":
		if contentType := mime.TypeByExtension("." + normalizedSuffix); contentType != "" {
			return contentType
		}
	}
	return "application/octet-stream"
}
