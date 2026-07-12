// This file contains upload handling, filename sanitization, duplicate hash
// reuse, and storage/database consistency cleanup for file records.

package file

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/grand"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	storagesvc "lina-core/internal/service/storage"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/closeutil"
	"lina-core/pkg/plugin/capability/storagecap"
)

// Upload handles file upload: computes SHA-256 hash, checks for duplicates, saves file via storage backend and records metadata in DB.
// If a file with the same hash already exists, a new record is still created (with different scene), reusing the physical file.
func (s *serviceImpl) Upload(ctx context.Context, in *UploadInput) (output *UploadOutput, err error) {
	if in == nil || in.File == nil {
		return nil, bizerr.NewCode(CodeFileUploadRequired)
	}
	file := in.File
	uploadMaxSize, err := s.configSvc.GetUploadMaxSize(ctx)
	if err != nil {
		return nil, err
	}
	if file.Size > uploadLimitBytes(uploadMaxSize) {
		return nil, bizerr.NewCode(CodeFileTooLarge, bizerr.P("maxSizeMB", uploadMaxSize))
	}
	src, err := file.Open()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeFileOpenFailed)
	}
	defer closeutil.Close(ctx, src, &err, "close uploaded file failed")
	return s.createFromReader(ctx, &CreateFromReaderInput{
		Filename:  file.Filename,
		Scene:     in.Scene,
		Reader:    src,
		SizeBytes: file.Size,
	}, uploadMaxSize)
}

// CreateFromReader creates a file-center metadata record from a generic stream.
func (s *serviceImpl) CreateFromReader(ctx context.Context, in *CreateFromReaderInput) (output *UploadOutput, err error) {
	uploadMaxSize, err := s.configSvc.GetUploadMaxSize(ctx)
	if err != nil {
		return nil, err
	}
	return s.createFromReader(ctx, in, uploadMaxSize)
}

// createFromReader owns shared file-center write semantics for HTTP uploads and plugin capabilities.
func (s *serviceImpl) createFromReader(ctx context.Context, in *CreateFromReaderInput, uploadMaxSize int64) (output *UploadOutput, err error) {
	if in == nil || in.Reader == nil {
		return nil, bizerr.NewCode(CodeFileUploadRequired)
	}
	sanitizedFilename := sanitizeFilename(in.Filename)
	if sanitizedFilename == "" || sanitizedFilename == "." {
		return nil, bizerr.NewCode(CodeFileUploadRequired)
	}
	if in.SizeBytes > uploadLimitBytes(uploadMaxSize) {
		return nil, bizerr.NewCode(CodeFileTooLarge, bizerr.P("maxSizeMB", uploadMaxSize))
	}
	// Hash-dedup reuses an existing object and skips storage.Put. Still require a
	// resolvable object backend so multi-cloud conflict fails closed for every
	// upload path (file and image, new content and duplicate content).
	if err = s.ensureFilesBackendReady(ctx); err != nil {
		return nil, err
	}
	spooled, actualSize, fileHash, err := spoolUploadReader(in.Reader, uploadMaxSize)
	if err != nil {
		return nil, err
	}
	defer cleanupSpooledUpload(spooled, &err)

	var userId int64
	if s.bizCtxSvc != nil {
		if bizCtx := s.bizCtxSvc.Get(ctx); bizCtx != nil {
			userId = int64(bizCtx.UserId)
		} else {
			userId = int64(s.bizCtxSvc.Current(ctx).UserID)
		}
	}
	tenantID := datascope.CurrentTenantID(ctx)
	scene := in.Scene
	if scene == "" {
		scene = DefaultFileSceneOther
	}
	suffix := gstr.ToLower(gfile.ExtName(sanitizedFilename))

	err = dao.SysFile.Ctx(ctx).Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		var existing *entity.SysFile
		existingModel := dao.SysFile.Ctx(ctx).
			Where(dao.SysFile.Columns().Hash, fileHash).
			Where(datascope.TenantColumn, tenantID)
		err := existingModel.Scan(&existing)
		if err != nil {
			return bizerr.WrapCode(err, CodeFileHashQueryFailed)
		}
		if existing != nil {
			result, err := dao.SysFile.Ctx(ctx).Data(do.SysFile{
				TenantId:  tenantID,
				Name:      existing.Name,
				Original:  sanitizedFilename,
				Suffix:    suffix,
				Scene:     scene,
				Size:      actualSize,
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
				Size:     actualSize,
			}
			return nil
		}
		if _, err = spooled.Seek(0, io.SeekStart); err != nil {
			return bizerr.WrapCode(err, CodeFileReadResetFailed)
		}
		storagePath := buildStorageKey(ctx, sanitizedFilename)
		// Pass a non-closing ReadSeeker so cloud SDKs (e.g. COS TeeReader) cannot
		// close the host-owned temp file before cleanupSpooledUpload runs.
		_, err = s.storage.Put(ctx, storagesvc.PutInput{
			Namespace: storagesvc.NamespaceFiles,
			Key:       storagePath,
			Body:      storageUploadBody{ReadSeeker: spooled},
			Size:      actualSize,
			Overwrite: true,
		})
		if err != nil {
			if bizerr.Is(err, storagecap.CodeStorageProviderConflict) {
				return bizerr.WrapCode(err, CodeFileStorageConflict)
			}
			return bizerr.WrapCode(err, CodeFileStoreFailed)
		}
		engine := storagesvc.FilesProviderID(ctx, s.storage)
		if strings.TrimSpace(engine) == "" {
			engine = EngineLocal
		}
		storedName := gfile.Basename(storagePath)
		url := storageURL(storagePath)
		result, err := dao.SysFile.Ctx(ctx).Data(do.SysFile{
			TenantId:  tenantID,
			Name:      storedName,
			Original:  sanitizedFilename,
			Suffix:    suffix,
			Scene:     scene,
			Size:      actualSize,
			Hash:      fileHash,
			Url:       url,
			Path:      storagePath,
			Engine:    engine,
			CreatedBy: userId,
		}).Insert()
		if err != nil {
			if cleanupErr := s.storage.Delete(ctx, storagesvc.DeleteInput{Namespace: storagesvc.NamespaceFiles, Key: storagePath}); cleanupErr != nil {
				return bizerr.WrapCode(
					fmt.Errorf("cleanup stored file after record save failure: %w; cleanup error: %w", err, cleanupErr),
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
			Size:     actualSize,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

// ensureFilesBackendReady fails closed when the host cannot select a single
// object backend for NamespaceFiles (for example multiple cloud provider
// plugins are enabled).
func (s *serviceImpl) ensureFilesBackendReady(ctx context.Context) error {
	if s == nil || s.storage == nil {
		return bizerr.NewCode(CodeFileStoreFailed)
	}
	if _, err := storagesvc.ResolveFilesProviderID(ctx, s.storage); err != nil {
		if bizerr.Is(err, storagecap.CodeStorageProviderConflict) {
			return bizerr.WrapCode(err, CodeFileStorageConflict)
		}
		return bizerr.WrapCode(err, CodeFileStoreFailed)
	}
	return nil
}

// buildStorageKey returns the file-center object key persisted in sys_file.path.
func buildStorageKey(ctx context.Context, filename string) string {
	var (
		now        = gtime.Now()
		dir        = fmt.Sprintf("%d/%s/%s", datascope.CurrentTenantID(ctx), now.Format("Y"), now.Format("m"))
		ext        = gfile.ExtName(filename)
		storedName = fmt.Sprintf("%s_%s", now.Format("Ymd_His"), grand.S(8))
	)
	if ext != "" {
		storedName += "." + gstr.ToLower(ext)
	}
	return gfile.Join(dir, storedName)
}

// storageURL returns the public file-center access URL for one object key.
func storageURL(storagePath string) string {
	return "/api/v1/uploads/" + storagePath
}

// uploadLimitBytes converts the configured MiB ceiling into bytes.
func uploadLimitBytes(maxSizeMB int64) int64 {
	return maxSizeMB * 1024 * 1024
}

// spoolUploadReader hashes a stream while copying it to a temporary seekable file.
func spoolUploadReader(reader io.Reader, uploadMaxSize int64) (*os.File, int64, string, error) {
	spooled, err := os.CreateTemp("", "lina-file-upload-*")
	if err != nil {
		return nil, 0, "", bizerr.WrapCode(err, CodeFileOpenFailed)
	}
	hasher := sha256.New()
	maxBytes := uploadLimitBytes(uploadMaxSize)
	written, err := io.Copy(io.MultiWriter(spooled, hasher), io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return nil, 0, "", cleanupSpooledUploadError(spooled, bizerr.WrapCode(err, CodeFileHashFailed))
	}
	if written > maxBytes {
		return nil, 0, "", cleanupSpooledUploadError(spooled, bizerr.NewCode(CodeFileTooLarge, bizerr.P("maxSizeMB", uploadMaxSize)))
	}
	if _, err = spooled.Seek(0, io.SeekStart); err != nil {
		return nil, 0, "", cleanupSpooledUploadError(spooled, bizerr.WrapCode(err, CodeFileReadResetFailed))
	}
	return spooled, written, hex.EncodeToString(hasher.Sum(nil)), nil
}

// storageUploadBody exposes only Read/Seek to storage backends.
// It intentionally omits Close so HTTP/cloud SDKs that close request bodies
// cannot take ownership of the host-managed spooled temp file.
type storageUploadBody struct {
	io.ReadSeeker
}

// cleanupSpooledUpload closes and removes the temporary upload file.
func cleanupSpooledUpload(spooled *os.File, errp *error) {
	if spooled == nil {
		return
	}
	name := spooled.Name()
	if closeErr := spooled.Close(); closeErr != nil && !errors.Is(closeErr, os.ErrClosed) && *errp == nil {
		*errp = bizerr.WrapCode(closeErr, CodeFileReadResetFailed)
	}
	if removeErr := os.Remove(name); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) && *errp == nil {
		*errp = bizerr.WrapCode(removeErr, CodeFileStoreFailed)
	}
}

// cleanupSpooledUploadError adds temporary-file cleanup failures to the primary error.
func cleanupSpooledUploadError(spooled *os.File, primary error) error {
	if spooled == nil {
		return primary
	}
	name := spooled.Name()
	if closeErr := spooled.Close(); closeErr != nil && !errors.Is(closeErr, os.ErrClosed) {
		return bizerr.WrapCode(fmt.Errorf("%w; cleanup close error: %w", primary, closeErr), CodeFileRecordSaveCleanupFailed)
	}
	if removeErr := os.Remove(name); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
		return bizerr.WrapCode(fmt.Errorf("%w; cleanup remove error: %w", primary, removeErr), CodeFileRecordSaveCleanupFailed)
	}
	return primary
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
