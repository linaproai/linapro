// Package file implements file upload, storage, download, and metadata query
// services for the Lina core host service.
package file

import (
	"context"
	"io"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	dictsvc "lina-core/internal/service/dict"
	storagesvc "lina-core/internal/service/storage"
)

// File storage engine and export limit constants.
const (
	EngineLocal   = "local" // Local storage engine identifier
	MaxExportRows = 10000   // Maximum rows for export
	// DefaultFileSceneOther is the dictionary value used when callers omit a file scene.
	DefaultFileSceneOther = "other"
)

// DictTypeFileScene is dict type used in file management
const DictTypeFileScene = "sys_file_scene"

// Service defines the file service contract.
type Service interface {
	// Upload stores one uploaded file after size validation, filename
	// sanitization, SHA-256 hashing, and duplicate physical-file detection. A
	// duplicate hash creates a new metadata record that reuses storage; database
	// and storage failures are wrapped as file business errors, and cleanup is
	// attempted when record creation fails after storage writes.
	Upload(ctx context.Context, in *UploadInput) (output *UploadOutput, err error)
	// CreateFromReader creates one host file-center record from a caller-owned
	// stream after the same size validation, filename sanitization, hash-based
	// deduplication, storage write, tenant attribution, and rollback cleanup used
	// by HTTP uploads. The reader is consumed during the call; nil readers or
	// filenames return file upload business errors.
	CreateFromReader(ctx context.Context, in *CreateFromReaderInput) (output *UploadOutput, err error)
	// List returns paginated file records visible to the current data scope,
	// with optional metadata filters, validated ordering, full URL projection,
	// and uploader display names.
	List(ctx context.Context, in *ListInput) (*ListOutput, error)
	// Info returns one file metadata record by ID after data-scope visibility
	// validation. Missing or out-of-scope records return file business errors.
	Info(ctx context.Context, id int64) (*entity.SysFile, error)
	// InfoByIds returns multiple file metadata records after all requested IDs
	// pass data-scope visibility checks. Returned URLs include the request base
	// URL when available.
	InfoByIds(ctx context.Context, ids []int64) ([]*entity.SysFile, error)
	// Delete soft-deletes visible file metadata rows and best-effort removes the
	// physical objects from storage. Visibility failures abort before mutation;
	// storage cleanup failures are logged and do not roll back soft deletion.
	Delete(ctx context.Context, ids []int64) error
	// OpenByID opens a stored file stream by metadata ID for download after
	// data-scope validation. Missing metadata or storage objects return file
	// business errors.
	OpenByID(ctx context.Context, id int64) (*OpenOutput, error)
	// OpenByPath opens a stored file stream by storage path for URL access after
	// path normalization and tenant visibility checks. Absolute and parent paths
	// are rejected before storage access.
	OpenByPath(ctx context.Context, storagePath string) (*OpenOutput, error)
	// UsageScenes returns all configured file usage scenes from the dictionary
	// service. Dictionary lookup errors are propagated.
	UsageScenes(ctx context.Context) ([]*UsageScenesOutput, error)
	// Suffixes returns distinct file suffixes visible to the current data scope.
	// Database or data-scope errors are returned.
	Suffixes(ctx context.Context) ([]*SuffixesOutput, error)
	// Detail returns visible file metadata with full URL, uploader name, and
	// dictionary-derived scene label. Missing or out-of-scope records return file
	// business errors.
	Detail(ctx context.Context, id int64) (*DetailOutput, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	configSvc config.Service // Configuration service
	storage   storagesvc.Service
	bizCtxSvc bizctx.Service  // Business context service
	dictSvc   dictsvc.Service // Dictionary service for scene labels
	scopeSvc  datascope.Service
}

// New creates and returns a new file service from explicit runtime-owned dependencies.
// storage is the host-wide Storage Service used for file-center content Put/Get/Delete.
func New(configSvc config.Service, storage storagesvc.Service, bizCtxSvc bizctx.Service, dictSvc dictsvc.Service, scopeSvc datascope.Service) Service {
	return &serviceImpl{
		configSvc: configSvc,
		storage:   storage,
		bizCtxSvc: bizCtxSvc,
		dictSvc:   dictSvc,
		scopeSvc:  scopeSvc,
	}
}

// UploadInput defines input for file upload.
type UploadInput struct {
	File  *ghttp.UploadFile // Uploaded file
	Scene string            // Usage scene
}

// CreateFromReaderInput defines input for creating a file from a stream.
type CreateFromReaderInput struct {
	Filename  string    // Original filename
	Scene     string    // Usage scene
	Reader    io.Reader // File content stream consumed during the call
	SizeBytes int64     // Optional known size in bytes; negative means unknown
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

// UsageScenesOutput defines output for usage scenes list.
type UsageScenesOutput struct {
	Value string `json:"value"` // Scene identifier
	Label string `json:"label"` // Scene name
}

// SuffixesOutput defines output for file suffix list.
type SuffixesOutput struct {
	Value string `json:"value"` // Suffix name
	Label string `json:"label"` // Display name
}

// DetailOutput defines output for file detail.
type DetailOutput struct {
	*entity.SysFile        // File entity
	CreatedByName   string `json:"createdByName"` // Uploader username
	SceneLabel      string `json:"sceneLabel"`    // Usage scene name
}
