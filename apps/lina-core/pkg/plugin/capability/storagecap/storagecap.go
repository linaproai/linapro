// Package storagecap defines the plugin-visible object storage capability and
// the provider extension contract. The Service contract owns plugin logical
// path semantics; Provider implementations only receive scoped object keys.
package storagecap

import (
	"context"
	"io"
	"time"
)

// Storage capability limits and stable provider identifiers.
const (
	// LocalProviderID identifies the host built-in local disk provider.
	LocalProviderID = "local"
	// VisibilityPrivate is the default plugin object visibility.
	VisibilityPrivate = "private"
	// DefaultListLimit bounds list results when callers omit a limit.
	DefaultListLimit = 100
	// MaxListLimit bounds list results for plugin-facing list calls.
	MaxListLimit = 1000
	// MaxLogicalPathBytes bounds plugin logical object path size.
	MaxLogicalPathBytes = 512
	// MaxBatchPathCount bounds one batch path request.
	MaxBatchPathCount = 100
	// MaxBatchPathBytes bounds the total logical path bytes in one batch request.
	MaxBatchPathBytes = MaxBatchPathCount * MaxLogicalPathBytes
)

// PutInput defines one plugin object write.
type PutInput struct {
	// Path is the plugin-local logical object path.
	Path string
	// Body carries the object content. Implementations read it during Put.
	Body io.Reader
	// Size optionally supplies body size. Negative means unknown.
	Size int64
	// ContentType is the optional MIME type.
	ContentType string
	// Overwrite controls whether an existing object may be replaced.
	Overwrite bool
}

// PutOutput defines one plugin object write result.
type PutOutput struct {
	// Object is the plugin-visible object metadata.
	Object *Object
}

// GetInput defines one plugin object read.
type GetInput struct {
	// Path is the plugin-local logical object path.
	Path string
}

// GetOutput defines one plugin object read result.
type GetOutput struct {
	// Object is the plugin-visible object metadata when Found is true.
	Object *Object
	// Body carries object content when Found is true. Callers must close it.
	Body io.ReadCloser
	// Found reports whether the object exists.
	Found bool
}

// DeleteInput defines one plugin object deletion.
type DeleteInput struct {
	// Path is the plugin-local logical object path.
	Path string
}

// DeleteManyInput defines one explicit-path batch deletion request. Missing
// objects are treated as successful no-ops by Service implementations.
type DeleteManyInput struct {
	// Paths are plugin-local logical object paths. Prefix deletes are not allowed.
	Paths []string
}

// ListInput defines one plugin object list request.
type ListInput struct {
	// Prefix is the plugin-local logical object prefix.
	Prefix string
	// Limit bounds the maximum returned objects. Zero uses DefaultListLimit.
	Limit int
}

// ListCursorInput defines one bounded cursor list request. Cursor values are
// opaque to plugins and only valid for the same Prefix.
type ListCursorInput struct {
	// Prefix is the plugin-local logical object prefix.
	Prefix string
	// Cursor resumes listing after the object represented by this cursor.
	Cursor string
	// Limit bounds the maximum returned objects. Zero uses DefaultListLimit.
	Limit int
}

// ListOutput defines one plugin object list response.
type ListOutput struct {
	// Objects contains plugin-visible object metadata.
	Objects []*Object
	// Limit is the effective limit applied by the service.
	Limit int
}

// ListCursorOutput defines one bounded cursor list response.
type ListCursorOutput struct {
	// Objects contains plugin-visible object metadata.
	Objects []*Object
	// NextCursor resumes the next page when non-empty.
	NextCursor string
	// Limit is the effective limit applied by the service.
	Limit int
}

// StatInput defines one plugin object metadata request.
type StatInput struct {
	// Path is the plugin-local logical object path.
	Path string
}

// StatOutput defines one plugin object metadata response.
type StatOutput struct {
	// Object is the plugin-visible object metadata when Found is true.
	Object *Object
	// Found reports whether the object exists.
	Found bool
}

// BatchStatInput defines one explicit-path metadata batch request.
type BatchStatInput struct {
	// Paths are plugin-local logical object paths. Prefix metadata reads are not allowed.
	Paths []string
}

// BatchStatOutput defines one metadata batch response. MissingPaths does not
// distinguish absent, invisible, or unauthorized paths.
type BatchStatOutput struct {
	// Objects contains metadata for visible existing objects.
	Objects []*Object
	// MissingPaths contains requested paths without a returned object.
	MissingPaths []string
}

// Object describes plugin-visible object metadata. It must never include
// provider object keys, local paths, credentials, or host file-management IDs.
type Object struct {
	// Path is the plugin-local logical object path.
	Path string
	// Size is object size in bytes.
	Size int64
	// ContentType is the normalized MIME type when known.
	ContentType string
	// ETag is provider metadata suitable for cache validation when available.
	ETag string
	// UpdatedAt is the provider object update timestamp when available.
	UpdatedAt *time.Time
	// Visibility describes plugin-visible object visibility.
	Visibility string
}

// Service defines plugin-scoped object storage operations. Implementations must
// scope every logical path by plugin ID and tenant context before delegating to
// a provider.
type Service interface {
	// Put writes one plugin object and returns metadata for the written object.
	Put(ctx context.Context, in PutInput) (*PutOutput, error)
	// Get reads one plugin object. A missing object returns Found=false.
	Get(ctx context.Context, in GetInput) (*GetOutput, error)
	// Delete removes one plugin object. Deleting a missing object is a no-op.
	Delete(ctx context.Context, in DeleteInput) error
	// DeleteMany removes an explicit bounded set of plugin objects. Missing objects are no-ops.
	DeleteMany(ctx context.Context, in DeleteManyInput) error
	// List lists plugin objects under one bounded logical prefix.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// ListCursor lists plugin objects under one bounded logical prefix with cursor pagination.
	ListCursor(ctx context.Context, in ListCursorInput) (*ListCursorOutput, error)
	// Stat reads plugin object metadata. A missing object returns Found=false.
	Stat(ctx context.Context, in StatInput) (*StatOutput, error)
	// BatchStat reads plugin object metadata for an explicit bounded path set.
	BatchStat(ctx context.Context, in BatchStatInput) (*BatchStatOutput, error)
	// ProviderStatuses returns registered provider status snapshots.
	ProviderStatuses(ctx context.Context) ([]*ProviderStatus, error)
}

// ProviderRuntime supplies provider factories with host runtime state needed to
// decide whether plugin-provided providers may serve requests.
type ProviderRuntime interface {
	// ProviderPluginAvailable reports whether a provider plugin is enabled and serviceable.
	ProviderPluginAvailable(ctx context.Context, pluginID string) bool
}

// ProviderEnv describes the environment passed to a provider factory.
type ProviderEnv struct {
	// ProviderID is the stable provider identifier being constructed.
	ProviderID string
	// Runtime exposes provider activation state without exposing host services.
	Runtime ProviderRuntime
}

// ProviderFactory constructs one provider instance for the current host runtime.
type ProviderFactory func(ctx context.Context, env ProviderEnv) (Provider, error)

// Provider defines the object-storage backend contract. Providers receive
// scoped object keys only and must not interpret plugin host-service
// authorization snapshots.
type Provider interface {
	// Put writes one object key and returns provider metadata.
	Put(ctx context.Context, in ProviderPutInput) (*ProviderObject, error)
	// Get reads one object key. A missing object returns Found=false.
	Get(ctx context.Context, in ProviderGetInput) (*ProviderGetOutput, error)
	// Delete removes one object key. Deleting a missing object is a no-op.
	Delete(ctx context.Context, in ProviderDeleteInput) error
	// DeleteMany removes explicit object keys. Deleting missing keys is a no-op.
	DeleteMany(ctx context.Context, in ProviderDeleteManyInput) error
	// List lists object keys under one bounded prefix.
	List(ctx context.Context, in ProviderListInput) (*ProviderListOutput, error)
	// ListCursor lists object keys under one bounded prefix with cursor pagination.
	ListCursor(ctx context.Context, in ProviderListCursorInput) (*ProviderListCursorOutput, error)
	// Stat reads one object key metadata. A missing object returns Found=false.
	Stat(ctx context.Context, in ProviderStatInput) (*ProviderStatOutput, error)
	// BatchStat reads object metadata for explicit object keys.
	BatchStat(ctx context.Context, in ProviderBatchStatInput) (*ProviderBatchStatOutput, error)
}

// ProviderPutInput defines one provider object write.
type ProviderPutInput struct {
	// Key is the scoped provider object key.
	Key string
	// Body carries the object content.
	Body io.Reader
	// Size optionally supplies body size. Negative means unknown.
	Size int64
	// ContentType is the normalized MIME type when known.
	ContentType string
	// Overwrite controls whether an existing object may be replaced.
	Overwrite bool
}

// ProviderGetInput defines one provider object read.
type ProviderGetInput struct {
	// Key is the scoped provider object key.
	Key string
}

// ProviderGetOutput defines one provider object read result.
type ProviderGetOutput struct {
	// Object is provider metadata when Found is true.
	Object *ProviderObject
	// Body carries object content when Found is true. Callers must close it.
	Body io.ReadCloser
	// Found reports whether the object exists.
	Found bool
}

// ProviderDeleteInput defines one provider object deletion.
type ProviderDeleteInput struct {
	// Key is the scoped provider object key.
	Key string
}

// ProviderDeleteManyInput defines one provider batch deletion request.
type ProviderDeleteManyInput struct {
	// Keys are scoped provider object keys.
	Keys []string
}

// ProviderListInput defines one provider object list request.
type ProviderListInput struct {
	// Prefix is the scoped provider object key prefix.
	Prefix string
	// Limit bounds returned objects.
	Limit int
}

// ProviderListCursorInput defines one provider cursor list request.
type ProviderListCursorInput struct {
	// Prefix is the scoped provider object key prefix.
	Prefix string
	// Cursor resumes listing after one provider object key.
	Cursor string
	// Limit bounds returned objects.
	Limit int
}

// ProviderListOutput defines one provider object list response.
type ProviderListOutput struct {
	// Objects contains provider object metadata.
	Objects []*ProviderObject
}

// ProviderListCursorOutput defines one provider cursor list response.
type ProviderListCursorOutput struct {
	// Objects contains provider object metadata.
	Objects []*ProviderObject
	// NextCursor resumes the next page when non-empty.
	NextCursor string
}

// ProviderStatInput defines one provider object metadata request.
type ProviderStatInput struct {
	// Key is the scoped provider object key.
	Key string
}

// ProviderStatOutput defines one provider object metadata response.
type ProviderStatOutput struct {
	// Object is provider metadata when Found is true.
	Object *ProviderObject
	// Found reports whether the object exists.
	Found bool
}

// ProviderBatchStatInput defines one provider metadata batch request.
type ProviderBatchStatInput struct {
	// Keys are scoped provider object keys.
	Keys []string
}

// ProviderBatchStatOutput defines one provider metadata batch response.
type ProviderBatchStatOutput struct {
	// Objects contains metadata for existing objects.
	Objects []*ProviderObject
	// MissingKeys contains requested keys without a returned object.
	MissingKeys []string
}

// ProviderObject describes provider-level object metadata. Key is never
// returned to plugins directly.
type ProviderObject struct {
	// Key is the scoped provider object key.
	Key string
	// Size is object size in bytes.
	Size int64
	// ContentType is the normalized MIME type when known.
	ContentType string
	// ETag is provider metadata suitable for cache validation when available.
	ETag string
	// UpdatedAt is the provider object update timestamp when available.
	UpdatedAt *time.Time
	// Visibility describes provider object visibility.
	Visibility string
}

// ProviderStatus describes one registered provider's activation state.
type ProviderStatus struct {
	// ProviderID identifies the provider. Built-in local provider uses LocalProviderID.
	ProviderID string
	// Active reports whether this provider currently receives storage calls.
	Active bool
	// Available reports whether this provider is usable.
	Available bool
	// Message carries a diagnostic string for unavailable providers.
	Message string
}
