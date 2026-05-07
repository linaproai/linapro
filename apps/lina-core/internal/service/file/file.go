// Package file implements file upload, storage, download, and metadata query
// services for the Lina core host service.
package file

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"path"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	dictsvc "lina-core/internal/service/dict"
	"lina-core/internal/service/orgcap"
	"lina-core/internal/service/role"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/closeutil"
	"lina-core/pkg/gdbutil"
	"lina-core/pkg/logger"
)

// File storage engine and export limit constants.
const (
	EngineLocal   = "local" // Local storage engine identifier
	MaxExportRows = 10000   // Maximum rows for export
)

// Dict type used in file management
const DictTypeFileScene = "sys_file_scene"

// Service defines the file service contract.
type Service interface {
	// Upload handles file upload: computes SHA-256 hash, checks for duplicates, saves file via storage backend and records metadata in DB.
	// If a file with the same hash already exists, a new record is still created (with different scene), reusing the physical file.
	Upload(ctx context.Context, in *UploadInput) (output *UploadOutput, err error)
	// List returns paginated file records.
	List(ctx context.Context, in *ListInput) (*ListOutput, error)
	// Info returns file info by ID.
	Info(ctx context.Context, id int64) (*entity.SysFile, error)
	// InfoByIds returns file info by multiple IDs.
	InfoByIds(ctx context.Context, ids []int64) ([]*entity.SysFile, error)
	// Delete removes files by IDs (soft delete in DB, also removes physical files).
	Delete(ctx context.Context, idsStr string) error
	// OpenByID opens a stored file stream by metadata ID for download.
	OpenByID(ctx context.Context, id int64) (*OpenOutput, error)
	// OpenByPath opens a stored file stream by metadata storage path for URL access.
	OpenByPath(ctx context.Context, storagePath string) (*OpenOutput, error)
	// UsageScenes returns all usage scenes from dictionary.
	UsageScenes(ctx context.Context) ([]*UsageScenesOutput, error)
	// Suffixes returns distinct file suffixes from the database.
	Suffixes(ctx context.Context) ([]*SuffixesOutput, error)
	// Detail returns file info with scene label.
	Detail(ctx context.Context, id int64) (*DetailOutput, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	configSvc config.Service  // Configuration service
	storage   Storage         // Storage backend
	bizCtxSvc bizctx.Service  // Business context service
	dictSvc   dictsvc.Service // Dictionary service for scene labels
	orgCapSvc orgcap.Service  // Optional organization capability service
	scopeSvc  datascope.Service
}

// New creates and returns a new Service instance with local storage.
func New(orgCapSvcs ...orgcap.Service) Service {
	var (
		ctx         = context.Background()
		configSvc   = config.New()
		storagePath = configSvc.GetUploadPath(ctx)
	)
	var orgCapSvc orgcap.Service
	if len(orgCapSvcs) > 0 {
		orgCapSvc = orgCapSvcs[0]
	}
	if orgCapSvc == nil {
		orgCapSvc = orgcap.New(nil)
	}
	svc := &serviceImpl{
		configSvc: configSvc,
		storage:   NewLocalStorage(storagePath),
		bizCtxSvc: bizctx.New(),
		dictSvc:   dictsvc.New(),
		orgCapSvc: orgCapSvc,
	}
	svc.scopeSvc = datascope.New(datascope.Dependencies{
		BizCtxSvc: svc.bizCtxSvc,
		RoleSvc:   role.New(nil),
		OrgCapSvc: svc.orgCapSvc,
	})
	return svc
}

// UploadInput defines input for file upload.
type UploadInput struct {
	File  *ghttp.UploadFile // Uploaded file
	Scene string            // Usage scene
}

// UploadOutput defines output for file upload.
type UploadOutput struct {
	Id       int64  `json:"id"`       // File ID
	Name     string `json:"name"`     // Stored filename
	Original string `json:"original"` // Original filename
	Url      string `json:"url"`      // File access URL
	Suffix   string `json:"suffix"`   // File suffix
	Size     int64  `json:"size"`     // File size (bytes)
}

// OpenOutput contains an opened file stream and response metadata.
type OpenOutput struct {
	Reader      io.ReadCloser // Reader streams file content from the configured storage backend
	Original    string        // Original filename
	Suffix      string        // File suffix
	ContentType string        // HTTP content type derived from file metadata
	Size        int64         // File size in bytes
}

// Upload handles file upload: computes SHA-256 hash, checks for duplicates, saves file via storage backend and records metadata in DB.
// If a file with the same hash already exists, a new record is still created (with different scene), reusing the physical file.
func (s *serviceImpl) Upload(ctx context.Context, in *UploadInput) (output *UploadOutput, err error) {
	file := in.File
	if file == nil {
		return nil, bizerr.NewCode(CodeFileUploadRequired)
	}

	// Sanitize filename to prevent path traversal attacks
	sanitizedFilename := sanitizeFilename(file.Filename)

	// Validate file size against the runtime-effective upload ceiling.
	uploadMaxSize, err := s.configSvc.GetUploadMaxSize(ctx)
	if err != nil {
		return nil, err
	}
	if file.Size > uploadMaxSize*1024*1024 {
		return nil, bizerr.NewCode(CodeFileTooLarge, bizerr.P("maxSizeMB", uploadMaxSize))
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeFileOpenFailed)
	}
	defer closeutil.Close(ctx, src, &err, "close uploaded file failed")

	// Compute SHA-256 hash
	hasher := sha256.New()
	if _, err = io.Copy(hasher, src); err != nil {
		return nil, bizerr.WrapCode(err, CodeFileHashFailed)
	}
	fileHash := hex.EncodeToString(hasher.Sum(nil))

	// Get current user ID
	var userId int64
	if bizCtx := s.bizCtxSvc.Get(ctx); bizCtx != nil {
		userId = int64(bizCtx.UserId)
	}

	scene := in.Scene
	if scene == "" {
		scene = "other"
	}
	suffix := gstr.ToLower(gfile.ExtName(sanitizedFilename))

	// Use transaction for database operations
	err = dao.SysFile.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Check for duplicate file by hash
		var existing *entity.SysFile
		err := dao.SysFile.Ctx(ctx).
			Where(dao.SysFile.Columns().Hash, fileHash).
			Scan(&existing)
		if err != nil {
			return bizerr.WrapCode(err, CodeFileHashQueryFailed)
		}

		if existing != nil {
			// Duplicate file found, reuse storage but create a new record with its own scene
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
			fullUrl := s.getBaseUrl(ctx) + existing.Url
			output = &UploadOutput{
				Id:       id,
				Name:     existing.Name,
				Original: sanitizedFilename,
				Url:      fullUrl,
				Suffix:   suffix,
				Size:     file.Size,
			}
			return nil
		}

		// Reset file reader position for storage
		if _, err = src.Seek(0, io.SeekStart); err != nil {
			return bizerr.WrapCode(err, CodeFileReadResetFailed)
		}

		// Save via storage backend
		storagePath, err := s.storage.Put(ctx, sanitizedFilename, src)
		if err != nil {
			return bizerr.WrapCode(err, CodeFileStoreFailed)
		}

		// Build file metadata
		storedName := gfile.Basename(storagePath)
		url := s.storage.Url(ctx, storagePath)

		// Insert file record
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
			// Clean up stored file on DB error
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
		// Return full URL with base URL prefix
		fullUrl := s.getBaseUrl(ctx) + url
		output = &UploadOutput{
			Id:       id,
			Name:     storedName,
			Original: sanitizedFilename,
			Url:      fullUrl,
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
	// Get the base name (remove any directory components)
	filename = filepath.Base(filename)

	// Remove null bytes and other control characters
	filename = strings.ReplaceAll(filename, "\x00", "")

	// Remove dangerous patterns
	dangerous := []string{"../", "..\\", "/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, d := range dangerous {
		filename = strings.ReplaceAll(filename, d, "_")
	}

	// Limit filename length
	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		name := filename[:255-len(ext)]
		filename = name + ext
	}

	return filename
}

// ListInput defines input for file list query.
type ListInput struct {
	PageNum        int    // Page number, starting from 1
	PageSize       int    // Page size
	Name           string // Stored filename, supports fuzzy search
	Original       string // Original filename, supports fuzzy search
	Suffix         string // File suffix
	Scene          string // Usage scene
	BeginTime      string // Creation time start
	EndTime        string // Creation time end
	OrderBy        string // Sort field
	OrderDirection string // Sort direction: asc/desc
}

// ListOutput defines output for file list.
type ListOutput struct {
	List  []*ListOutputItem `json:"list"`  // File list
	Total int               `json:"total"` // Total count
}

// ListOutputItem defines a single file item in list output.
type ListOutputItem struct {
	*entity.SysFile        // File entity
	CreatedByName   string `json:"createdByName"` // Uploader username
}

// List returns paginated file records.
func (s *serviceImpl) List(ctx context.Context, in *ListInput) (*ListOutput, error) {
	m := dao.SysFile.Ctx(ctx)

	if in.Name != "" {
		m = m.WhereLike(dao.SysFile.Columns().Name, fmt.Sprintf("%%%s%%", in.Name))
	}
	if in.Original != "" {
		m = m.WhereLike(dao.SysFile.Columns().Original, fmt.Sprintf("%%%s%%", in.Original))
	}
	if in.Suffix != "" {
		m = m.Where(dao.SysFile.Columns().Suffix, in.Suffix)
	}
	if in.BeginTime != "" {
		m = m.WhereGTE(dao.SysFile.Columns().CreatedAt, in.BeginTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.SysFile.Columns().CreatedAt, in.EndTime)
	}
	if in.Scene != "" {
		m = m.Where(dao.SysFile.Columns().Scene, in.Scene)
	}
	var err error
	m, err = s.applyFileDataScope(ctx, m)
	if err != nil {
		return nil, err
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	cols := dao.SysFile.Columns()
	var (
		orderBy           = cols.Id
		allowedSortFields = map[string]string{
			"size":      cols.Size,
			"createdAt": cols.CreatedAt,
		}
		direction = gdbutil.NormalizeOrderDirectionOrDefault(in.OrderDirection, gdbutil.OrderDirectionDESC)
	)
	if in.OrderBy != "" {
		if field, ok := allowedSortFields[in.OrderBy]; ok {
			orderBy = field
		}
	}

	var files []*entity.SysFile
	err = gdbutil.ApplyModelOrder(
		m.Page(in.PageNum, in.PageSize),
		orderBy,
		direction,
	).Scan(&files)
	if err != nil {
		return nil, err
	}

	// Collect unique creator user IDs for name resolution
	userIdMap := make(map[int64]bool)
	for _, f := range files {
		if f.CreatedBy > 0 {
			userIdMap[f.CreatedBy] = true
		}
	}
	userNameMap := make(map[int64]string)
	if len(userIdMap) > 0 {
		userIds := make([]int64, 0, len(userIdMap))
		for uid := range userIdMap {
			userIds = append(userIds, uid)
		}
		var users []*entity.SysUser
		err = dao.SysUser.Ctx(ctx).
			WhereIn(dao.SysUser.Columns().Id, userIds).
			Scan(&users)
		if err == nil {
			for _, u := range users {
				userNameMap[int64(u.Id)] = u.Username
			}
		}
	}

	// Build full URL prefix from HTTP request context
	baseUrl := s.getBaseUrl(ctx)

	items := make([]*ListOutputItem, len(files))
	for i, f := range files {
		fileCopy := *f
		if fileCopy.Url != "" && baseUrl != "" {
			fileCopy.Url = baseUrl + fileCopy.Url
		}
		items[i] = &ListOutputItem{
			SysFile:       &fileCopy,
			CreatedByName: userNameMap[f.CreatedBy],
		}
	}

	return &ListOutput{
		List:  items,
		Total: total,
	}, nil
}

// Info returns file info by ID.
func (s *serviceImpl) Info(ctx context.Context, id int64) (*entity.SysFile, error) {
	if err := s.ensureFilesVisible(ctx, []int64{id}); err != nil {
		return nil, err
	}
	var file *entity.SysFile
	err := dao.SysFile.Ctx(ctx).Where(dao.SysFile.Columns().Id, id).Scan(&file)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeFileRecordQueryFailed)
	}
	if file == nil {
		return nil, bizerr.NewCode(CodeFileNotFound)
	}
	return file, nil
}

// InfoByIds returns file info by multiple IDs.
func (s *serviceImpl) InfoByIds(ctx context.Context, ids []int64) ([]*entity.SysFile, error) {
	if err := s.ensureFilesVisible(ctx, ids); err != nil {
		return nil, err
	}
	var files []*entity.SysFile
	err := dao.SysFile.Ctx(ctx).WhereIn(dao.SysFile.Columns().Id, ids).Scan(&files)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeFileRecordQueryFailed)
	}
	// Build full URL prefix from HTTP request context
	baseUrl := s.getBaseUrl(ctx)
	if baseUrl != "" {
		for _, f := range files {
			if f.Url != "" {
				f.Url = baseUrl + f.Url
			}
		}
	}
	return files, nil
}

// Delete removes files by IDs (soft delete in DB, also removes physical files).
func (s *serviceImpl) Delete(ctx context.Context, idsStr string) error {
	ids := gstr.SplitAndTrim(idsStr, ",")
	if len(ids) == 0 {
		return bizerr.NewCode(CodeFileDeleteRequired)
	}

	idList := make([]int64, 0, len(ids))
	for _, idStr := range ids {
		idList = append(idList, gconv.Int64(idStr))
	}
	if err := s.ensureFilesVisible(ctx, idList); err != nil {
		return err
	}

	// Get file records first to delete physical files
	var files []*entity.SysFile
	err := dao.SysFile.Ctx(ctx).WhereIn(dao.SysFile.Columns().Id, idList).Scan(&files)
	if err != nil {
		return err
	}

	// Soft delete from DB
	_, err = dao.SysFile.Ctx(ctx).WhereIn(dao.SysFile.Columns().Id, idList).Delete()
	if err != nil {
		return err
	}

	// Delete physical files (best effort, don't fail on cleanup errors)
	for _, f := range files {
		if deleteErr := s.storage.Delete(ctx, f.Path); deleteErr != nil {
			logger.Warningf(ctx, "delete storage file failed path=%s err=%v", f.Path, deleteErr)
		}
	}

	return nil
}

// getBaseUrl returns the base URL (scheme + host) from the current HTTP request context.
func (s *serviceImpl) getBaseUrl(ctx context.Context) string {
	r := g.RequestFromCtx(ctx)
	if r == nil {
		return ""
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
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
	err := dao.SysFile.Ctx(ctx).
		Where(dao.SysFile.Columns().Path, normalizedPath).
		Scan(&fileInfo)
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

// UsageScenesOutput defines output for usage scenes list.
type UsageScenesOutput struct {
	Value string `json:"value"` // Scene identifier
	Label string `json:"label"` // Scene name
}

// UsageScenes returns all usage scenes from dictionary.
func (s *serviceImpl) UsageScenes(ctx context.Context) ([]*UsageScenesOutput, error) {
	list, err := s.dictSvc.DataByType(ctx, DictTypeFileScene)
	if err != nil {
		return nil, err
	}
	items := make([]*UsageScenesOutput, 0, len(list))
	for _, item := range list {
		items = append(items, &UsageScenesOutput{
			Value: item.Value,
			Label: item.Label,
		})
	}
	return items, nil
}

// SuffixesOutput defines output for file suffix list.
type SuffixesOutput struct {
	Value string `json:"value"` // Suffix name
	Label string `json:"label"` // Display name
}

// Suffixes returns distinct file suffixes from the database.
func (s *serviceImpl) Suffixes(ctx context.Context) ([]*SuffixesOutput, error) {
	model, err := s.applyFileDataScope(ctx, dao.SysFile.Ctx(ctx))
	if err != nil {
		return nil, err
	}
	result, err := model.Fields(dao.SysFile.Columns().Suffix).
		Group(dao.SysFile.Columns().Suffix).
		OrderAsc(dao.SysFile.Columns().Suffix).
		Array()
	if err != nil {
		return nil, err
	}
	items := make([]*SuffixesOutput, 0, len(result))
	for _, v := range result {
		suffix := v.String()
		if suffix == "" {
			continue
		}
		items = append(items, &SuffixesOutput{
			Value: suffix,
			Label: suffix,
		})
	}
	return items, nil
}

// DetailOutput defines output for file detail.
type DetailOutput struct {
	*entity.SysFile        // File entity
	CreatedByName   string `json:"createdByName"` // Uploader username
	SceneLabel      string `json:"sceneLabel"`    // Usage scene name
}

// Detail returns file info with scene label.
func (s *serviceImpl) Detail(ctx context.Context, id int64) (*DetailOutput, error) {
	if err := s.ensureFilesVisible(ctx, []int64{id}); err != nil {
		return nil, err
	}
	// Get file info
	var file *entity.SysFile
	err := dao.SysFile.Ctx(ctx).Where(dao.SysFile.Columns().Id, id).Scan(&file)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, bizerr.NewCode(CodeFileNotFound)
	}

	// Build full URL
	baseUrl := s.getBaseUrl(ctx)
	if baseUrl != "" && file.Url != "" {
		file.Url = baseUrl + file.Url
	}

	// Get uploader name
	var createdByName string
	if file.CreatedBy > 0 {
		var user *entity.SysUser
		err = dao.SysUser.Ctx(ctx).
			Where(dao.SysUser.Columns().Id, file.CreatedBy).
			Scan(&user)
		if err == nil && user != nil {
			createdByName = user.Username
		}
	}

	// Get scene label from dictionary
	sceneLabel := s.dictSvc.GetLabelByValue(ctx, dictsvc.GetLabelByValueInput{
		DictType: DictTypeFileScene,
		Value:    file.Scene,
	})

	return &DetailOutput{
		SysFile:       file,
		CreatedByName: createdByName,
		SceneLabel:    sceneLabel,
	}, nil
}
