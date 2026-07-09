// Package filecap defines governed file-domain capability contracts for
// plugins without exposing host file storage tables or physical paths.
package filecap

import (
	"context"
	"io"
	"lina-core/pkg/plugin/capability/capmodel"
)

// Service defines governed file capability methods. Reads use bounded
// info queries and tenant/data-scope filtering; deletes must validate
// target visibility and host file-center ownership before mutation.
type Service interface {
	// Get returns one visible file info record. Risk: read; resource: file ID;
	// context: actor and tenant; data permission: target visibility;
	// performance: delegates to BatchGet; audit/cache: read-only.
	Get(ctx context.Context, id FileID) (*FileInfo, error)
	// Detail returns one visible file detail without storage internals.
	Detail(ctx context.Context, id FileID) (*DetailInfo, error)
	// BatchGet returns visible file info records and opaque missing IDs.
	BatchGet(ctx context.Context, ids []FileID) (*capmodel.BatchResult[*FileInfo, FileID], error)
	// List returns one bounded page of visible file info records.
	List(ctx context.Context, input ListInput) (*capmodel.PageResult[*FileInfo], error)
	// ListScenes returns bounded governed file scene options.
	ListScenes(ctx context.Context) ([]*Option, error)
	// ListSuffixes returns bounded visible file suffix options.
	ListSuffixes(ctx context.Context) ([]*Option, error)
	// Open opens a visible file stream after target and tenant checks.
	Open(ctx context.Context, id FileID) (*FileContent, error)
	// EnsureVisible rejects when any requested file is absent or invisible.
	EnsureVisible(ctx context.Context, ids []FileID) error
	// Upload creates one host file record from uploaded content.
	Upload(ctx context.Context, input UploadInput) (*FileInfo, error)
	// CreateFromStorage creates one host file record from a plugin storage object.
	CreateFromStorage(ctx context.Context, input CreateFromStorageInput) (*FileInfo, error)
	// UpdateMetadata mutates governed visible file metadata.
	UpdateMetadata(ctx context.Context, input UpdateMetadataInput) error
	// Delete deletes visible files after target, tenant, scene, data-scope,
	// audit, and host file-center boundary checks.
	Delete(ctx context.Context, id FileID) error
	// DeleteMany deletes visible files after target, tenant, scene, data-scope,
	// audit, and host file-center boundary checks.
	DeleteMany(ctx context.Context, ids []FileID) error
}

const (
	// MaxBatchGetFiles limits one batch file projection request.
	MaxBatchGetFiles = 200
	// MaxEnsureVisibleFiles limits one target visibility check.
	MaxEnsureVisibleFiles = MaxBatchGetFiles
	// MaxListPageSize limits one file candidate list page.
	MaxListPageSize = 200
	// MaxDirectUploadBytes bounds dynamic-plugin files.upload JSON payloads.
	MaxDirectUploadBytes int64 = 10 * 1024 * 1024
)

// FileID identifies one governed file resource.
type FileID string

// FileInfo describes one file reference visible to a plugin.
type FileInfo struct {
	// ID is the governed file identifier.
	ID FileID
	// Name is the original or display file name.
	Name string
	// MimeType is the media type.
	MimeType string
	// SizeBytes is the file size.
	SizeBytes int64
	// BusinessScene is the domain-owned business scene.
	BusinessScene string
}

// DetailInfo describes a visible file detail without storage internals.
type DetailInfo struct {
	// FileInfo is the stable file summary information.
	FileInfo
	// OriginalName is the original uploaded filename.
	OriginalName string
	// URL is a governed access URL when the owner can expose one.
	URL string
	// CreatedByName is the uploader display name when visible.
	CreatedByName string
	// SceneLabel is the current-locale scene label when available.
	SceneLabel string
}

// FileContent describes an opened governed file stream.
type FileContent struct {
	// Reader streams file content. Callers must close it.
	Reader io.ReadCloser
	// Filename is the response filename.
	Filename string
	// ContentType is the response content type.
	ContentType string
	// SizeBytes is the content length when known.
	SizeBytes int64
}

// Option describes a small file metadata option such as scene or suffix.
type Option struct {
	// Value is the stable option value.
	Value string
	// Label is the display label.
	Label string
	// LabelKey is the optional runtime i18n key.
	LabelKey string
}

// ListInput constrains governed file candidate listing.
type ListInput struct {
	// BusinessScene filters by host file usage scene.
	BusinessScene string
	// Keyword filters by original or stored file name.
	Keyword string
	// MimeType filters by coarse MIME type inferred from file suffix.
	MimeType string
	// Page constrains result size and stable sorting.
	Page capmodel.PageRequest
}

// UploadInput describes one governed file upload. Dynamic transport can use
// CreateFromStorage for plugin-private object promotion when streaming upload
// is not available.
type UploadInput struct {
	// Filename is the original filename.
	Filename string
	// BusinessScene is the governed host file usage scene.
	BusinessScene string
	// Reader streams the file content.
	Reader io.Reader
	// SizeBytes is the upload size when known.
	SizeBytes int64
}

// CreateFromStorageInput describes creation of a host file record from a
// plugin-private storage object.
type CreateFromStorageInput struct {
	// StoragePath is the plugin-private storage path.
	StoragePath string
	// Filename is the display filename.
	Filename string
	// BusinessScene is the governed host file usage scene.
	BusinessScene string
	// SizeBytes is the object size when known.
	SizeBytes int64
}

// UpdateMetadataInput describes visible file metadata changes.
type UpdateMetadataInput struct {
	// ID identifies the target file.
	ID FileID
	// Name optionally updates the display filename.
	Name *string
	// BusinessScene optionally updates the host usage scene.
	BusinessScene *string
}
