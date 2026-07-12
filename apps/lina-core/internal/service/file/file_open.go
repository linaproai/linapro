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

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	storagesvc "lina-core/internal/service/storage"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
)

// fileSuffixesByMimeType maps stable MIME filters to storage suffixes.
var fileSuffixesByMimeType = map[string][]string{
	"image/jpeg":       {"jpg", "jpeg"},
	"image/png":        {"png"},
	"image/gif":        {"gif"},
	"image/webp":       {"webp"},
	"image/svg+xml":    {"svg"},
	"application/pdf":  {"pdf"},
	"text/plain":       {"txt", "log"},
	"application/json": {"json"},
}

// browserContentTypeSuffixes lists suffixes allowed to use browser content types.
var browserContentTypeSuffixes = map[string]struct{}{
	"gif":  {},
	"jpeg": {},
	"jpg":  {},
	"pdf":  {},
	"png":  {},
	"svg":  {},
	"webp": {},
}

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
	output, err := s.storage.Get(ctx, storagesvc.GetInput{Namespace: storagesvc.NamespaceFiles, Key: fileInfo.Path})
	if err != nil {
		if bizerr.Is(err, storagecap.CodeStorageProviderConflict) {
			return nil, bizerr.WrapCode(err, CodeFileStorageConflict)
		}
		if errors.Is(err, fs.ErrNotExist) || errors.Is(err, storagesvc.ErrPathInvalid) {
			return nil, bizerr.NewCode(CodeFileNotFound)
		}
		return nil, bizerr.WrapCode(err, CodeFileStorageReadFailed)
	}
	if output == nil || !output.Found || output.Body == nil {
		return nil, bizerr.NewCode(CodeFileNotFound)
	}
	return &OpenOutput{
		Reader:      output.Body,
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
	normalizedSuffix := normalizeFileSuffix(suffix)
	if _, ok := browserContentTypeSuffixes[normalizedSuffix]; !ok {
		return "application/octet-stream"
	}
	if contentType := mime.TypeByExtension("." + normalizedSuffix); contentType != "" {
		return contentType
	}
	return mimeTypeFromSuffix(normalizedSuffix)
}

// suffixesForMimeType returns storage suffixes matching a stable coarse media type.
func suffixesForMimeType(mimeType string) []string {
	suffixes := fileSuffixesByMimeType[strings.ToLower(strings.TrimSpace(mimeType))]
	if len(suffixes) == 0 {
		return nil
	}
	result := make([]string, len(suffixes))
	copy(result, suffixes)
	return result
}

// mimeTypeFromSuffix returns a stable coarse media type for plugin projections.
func mimeTypeFromSuffix(suffix string) string {
	normalizedSuffix := normalizeFileSuffix(suffix)
	for mimeType, suffixes := range fileSuffixesByMimeType {
		for _, candidate := range suffixes {
			if candidate == normalizedSuffix {
				return mimeType
			}
		}
	}
	return ""
}

// normalizeFileSuffix returns a lowercase suffix without a leading dot.
func normalizeFileSuffix(suffix string) string {
	return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(suffix)), ".")
}
