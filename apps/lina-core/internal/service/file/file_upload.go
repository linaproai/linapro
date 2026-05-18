// This file contains upload handling, filename sanitization, duplicate hash
// reuse, and storage/database consistency cleanup for file records.

package file

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/closeutil"
)

// Upload handles file upload: computes SHA-256 hash, checks for duplicates, saves file via storage backend and records metadata in DB.
// If a file with the same hash already exists, a new record is still created (with different scene), reusing the physical file.
func (s *serviceImpl) Upload(ctx context.Context, in *UploadInput) (output *UploadOutput, err error) {
	file := in.File
	if file == nil {
		return nil, bizerr.NewCode(CodeFileUploadRequired)
	}
	sanitizedFilename := sanitizeFilename(file.Filename)
	uploadMaxSize, err := s.configSvc.GetUploadMaxSize(ctx)
	if err != nil {
		return nil, err
	}
	if file.Size > uploadMaxSize*1024*1024 {
		return nil, bizerr.NewCode(CodeFileTooLarge, bizerr.P("maxSizeMB", uploadMaxSize))
	}
	src, err := file.Open()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeFileOpenFailed)
	}
	defer closeutil.Close(ctx, src, &err, "close uploaded file failed")

	hasher := sha256.New()
	if _, err = io.Copy(hasher, src); err != nil {
		return nil, bizerr.WrapCode(err, CodeFileHashFailed)
	}
	fileHash := hex.EncodeToString(hasher.Sum(nil))

	var userId int64
	if bizCtx := s.bizCtxSvc.Get(ctx); bizCtx != nil {
		userId = int64(bizCtx.UserId)
	}
	scene := in.Scene
	if scene == "" {
		scene = "other"
	}
	suffix := gstr.ToLower(gfile.ExtName(sanitizedFilename))

	err = dao.SysFile.Ctx(ctx).Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		var existing *entity.SysFile
		existingModel := dao.SysFile.Ctx(ctx).Where(dao.SysFile.Columns().Hash, fileHash)
		existingModel = datascope.ApplyTenantScope(ctx, existingModel, datascope.TenantColumn)
		err := existingModel.Scan(&existing)
		if err != nil {
			return bizerr.WrapCode(err, CodeFileHashQueryFailed)
		}
		if existing != nil {
			result, err := dao.SysFile.Ctx(ctx).Data(do.SysFile{
				Name:      existing.Name,
				Original:  sanitizedFilename,
				Suffix:    suffix,
				Scene:     scene,
				Size:      file.Size,
				Hash:      fileHash,
				Url:       existing.Url,
				Path:      existing.Path,
				Engine:    existing.Engine,
				CreatedBy: userId,
			}).Insert()
			if err != nil {
				return bizerr.WrapCode(err, CodeFileRecordSaveFailed)
			}
			id, err := result.LastInsertId()
			if err != nil {
				return bizerr.WrapCode(err, CodeFileRecordIDReadFailed)
			}
			output = &UploadOutput{
				Id:       id,
				Name:     existing.Name,
				Original: sanitizedFilename,
				Url:      s.getBaseUrl(ctx) + existing.Url,
				Suffix:   suffix,
				Size:     file.Size,
			}
			return nil
		}
		if _, err = src.Seek(0, io.SeekStart); err != nil {
			return bizerr.WrapCode(err, CodeFileReadResetFailed)
		}
		storagePath, err := s.storage.Put(ctx, sanitizedFilename, src)
		if err != nil {
			return bizerr.WrapCode(err, CodeFileStoreFailed)
		}
		storedName := gfile.Basename(storagePath)
		url := s.storage.Url(ctx, storagePath)
		result, err := dao.SysFile.Ctx(ctx).Data(do.SysFile{
			Name:      storedName,
			Original:  sanitizedFilename,
			Suffix:    suffix,
			Scene:     scene,
			Size:      file.Size,
			Hash:      fileHash,
			Url:       url,
			Path:      storagePath,
			Engine:    EngineLocal,
			CreatedBy: userId,
		}).Insert()
		if err != nil {
			if cleanupErr := s.storage.Delete(ctx, storagePath); cleanupErr != nil {
				return bizerr.WrapCode(
					fmt.Errorf("cleanup stored file after record save failure: %w; cleanup error: %v", err, cleanupErr),
					CodeFileRecordSaveCleanupFailed,
				)
			}
			return bizerr.WrapCode(err, CodeFileRecordSaveFailed)
		}
		id, err := result.LastInsertId()
		if err != nil {
			return bizerr.WrapCode(err, CodeFileRecordIDReadFailed)
		}
		output = &UploadOutput{
			Id:       id,
			Name:     storedName,
			Original: sanitizedFilename,
			Url:      s.getBaseUrl(ctx) + url,
			Suffix:   suffix,
			Size:     file.Size,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

// sanitizeFilename removes path traversal characters and dangerous patterns from filename.
func sanitizeFilename(filename string) string {
	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, "\x00", "")
	dangerous := []string{"../", "..\\", "/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, d := range dangerous {
		filename = strings.ReplaceAll(filename, d, "_")
	}
	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		name := filename[:255-len(ext)]
		filename = name + ext
	}
	return filename
}
