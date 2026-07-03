// Package storage implements host-owned object storage primitives shared by
// file-center and plugin object-storage domains without owning their metadata,
// authorization, URL, or public capability semantics.
package storage

import (
	"context"
	"errors"
	"io"
	"time"
)

// Stable storage namespaces owned by the host.
const (
	// NamespaceFiles stores host file-center objects.
	NamespaceFiles = "files"
	// NamespacePlugins stores plugin provider objects.
	NamespacePlugins = "plugins"
	// DefaultListLimit is used when callers omit a positive list limit.
	DefaultListLimit = 100
	// MaxListLimit caps local object list responses.
	MaxListLimit = 1000
)

// Sentinel errors returned by the internal object storage boundary. Domain
// callers translate these errors to their own business error codes.
var (
	ErrUnavailable = errors.New("storage unavailable")
	ErrPathInvalid = errors.New("storage path invalid")
	ErrObjectExist = errors.New("storage object exists")
)

// Service defines the host-internal object storage contract. It owns physical
// object persistence only; callers own namespace selection, domain metadata,
// data permission checks, URL generation, and plugin-visible path semantics.
type Service interface {
	// Put writes one object key under a namespace and returns object metadata.
	// Existing objects return ErrObjectExist when Overwrite is false. Nil bodies
	// are written as empty objects; empty namespace or unsafe keys return
	// ErrPathInvalid, and missing namespace roots return ErrUnavailable.
	Put(ctx context.Context, in PutInput) (*PutOutput, error)
	// Get opens one object key under a namespace and returns a caller-owned
	// stream when Found is true. Missing objects return Found=false with nil
	// Body; invalid namespace/key values return ErrPathInvalid, and missing
	// namespace roots return ErrUnavailable.
	Get(ctx context.Context, in GetInput) (*GetOutput, error)
	// Delete removes one object key under a namespace. Missing objects are
	// successful no-ops; invalid namespace/key values return ErrPathInvalid, and
	// missing namespace roots return ErrUnavailable.
	Delete(ctx context.Context, in DeleteInput) error
	// DeleteMany removes an explicit bounded set of object keys under a
	// namespace. Nil or empty key slices are no-ops; the first invalid key or
	// storage failure is returned.
	DeleteMany(ctx context.Context, in DeleteManyInput) error
	// List returns objects under one namespace prefix using an effective limit.
	// Empty prefixes list from the namespace root, zero limits use
	// DefaultListLimit, oversized limits are capped, and missing prefixes return
	// an empty object slice.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// ListCursor returns objects under one namespace prefix after an optional
	// cursor. Cursor values are normalized object keys; empty cursors start at
	// the first object, and NextCursor is empty when no further page is known.
	ListCursor(ctx context.Context, in ListCursorInput) (*ListCursorOutput, error)
	// Stat returns metadata for one object key. Missing objects return
	// Found=false with nil Object; invalid namespace/key values return
	// ErrPathInvalid, and missing namespace roots return ErrUnavailable.
	Stat(ctx context.Context, in StatInput) (*StatOutput, error)
	// BatchStat returns metadata for an explicit bounded set of object keys. Nil
	// or empty key slices return empty Objects and MissingKeys; missing objects
	// are reported in MissingKeys without exposing whether a domain caller may
	// see the object.
	BatchStat(ctx context.Context, in BatchStatInput) (*BatchStatOutput, error)
}

// Config defines local object-storage roots. NamespaceRoots override RootDir
// for specific namespaces so callers can preserve existing runtime paths.
type Config struct {
	// RootDir is the fallback local root used when a namespace has no override.
	RootDir string
	// NamespaceRoots maps namespace names to local roots and takes precedence over RootDir.
	NamespaceRoots map[string]string
}

// serviceImpl implements Service using local filesystem roots.
type serviceImpl struct {
	// rootDir is the fallback local filesystem root.
	rootDir string
	// namespaceRoots contains namespace-specific local filesystem roots.
	namespaceRoots map[string]string
}

var _ Service = (*serviceImpl)(nil)

// New creates a local filesystem-backed host object storage service.
func New(config Config) Service {
	roots := make(map[string]string, len(config.NamespaceRoots))
	for namespace, root := range config.NamespaceRoots {
		roots[namespace] = root
	}
	return &serviceImpl{
		rootDir:        config.RootDir,
		namespaceRoots: roots,
	}
}

// PutInput defines one object write.
type PutInput struct {
	// Namespace selects the storage root.
	Namespace string
	// Key is a relative object key inside Namespace.
	Key string
	// Body is the stream consumed during Put; nil means an empty object.
	Body io.Reader
	// Size optionally carries caller-known object size; the local backend records actual size.
	Size int64
	// ContentType optionally carries caller-known MIME metadata.
	ContentType string
	// Overwrite allows replacing an existing object when true.
	Overwrite bool
}

// PutOutput defines one object write result.
type PutOutput struct {
	// Object contains metadata for the written object.
	Object *Object
}

// GetInput defines one object read.
type GetInput struct {
	// Namespace selects the storage root.
	Namespace string
	// Key is a relative object key inside Namespace.
	Key string
}

// GetOutput defines one object read result.
type GetOutput struct {
	// Object contains metadata when Found is true.
	Object *Object
	// Body is the caller-owned stream when Found is true.
	Body io.ReadCloser
	// Found reports whether the object exists.
	Found bool
}

// DeleteInput defines one object deletion.
type DeleteInput struct {
	// Namespace selects the storage root.
	Namespace string
	// Key is a relative object key inside Namespace.
	Key string
}

// DeleteManyInput defines one explicit-key batch deletion.
type DeleteManyInput struct {
	// Namespace selects the storage root.
	Namespace string
	// Keys are explicit relative object keys inside Namespace.
	Keys []string
}

// ListInput defines one bounded prefix listing.
type ListInput struct {
	// Namespace selects the storage root.
	Namespace string
	// Prefix is a relative prefix inside Namespace; empty lists from the root.
	Prefix string
	// Limit bounds returned objects; zero uses DefaultListLimit.
	Limit int
}

// ListOutput defines one bounded prefix listing result.
type ListOutput struct {
	// Objects contains matching object metadata.
	Objects []*Object
	// Limit is the effective limit used by the backend.
	Limit int
}

// ListCursorInput defines one bounded cursor listing.
type ListCursorInput struct {
	// Namespace selects the storage root.
	Namespace string
	// Prefix is a relative prefix inside Namespace; empty lists from the root.
	Prefix string
	// Cursor resumes after a relative object key from a previous page.
	Cursor string
	// Limit bounds returned objects; zero uses DefaultListLimit.
	Limit int
}

// ListCursorOutput defines one bounded cursor listing result.
type ListCursorOutput struct {
	// Objects contains matching object metadata for the current page.
	Objects []*Object
	// NextCursor resumes the next page when non-empty.
	NextCursor string
	// Limit is the effective limit used by the backend.
	Limit int
}

// StatInput defines one object metadata read.
type StatInput struct {
	// Namespace selects the storage root.
	Namespace string
	// Key is a relative object key inside Namespace.
	Key string
}

// StatOutput defines one object metadata read result.
type StatOutput struct {
	// Object contains metadata when Found is true.
	Object *Object
	// Found reports whether the object exists.
	Found bool
}

// BatchStatInput defines one explicit-key batch metadata read.
type BatchStatInput struct {
	// Namespace selects the storage root.
	Namespace string
	// Keys are explicit relative object keys inside Namespace.
	Keys []string
}

// BatchStatOutput defines one explicit-key batch metadata read result.
type BatchStatOutput struct {
	// Objects contains metadata for found objects.
	Objects []*Object
	// MissingKeys contains requested keys without returned metadata.
	MissingKeys []string
}

// Object contains host-internal object metadata without domain identifiers,
// public URLs, provider credentials, or local absolute paths.
type Object struct {
	// Key is the normalized relative object key.
	Key string
	// Size is the object size in bytes.
	Size int64
	// ContentType is caller-supplied MIME metadata when known.
	ContentType string
	// ETag is backend metadata suitable for cache validation.
	ETag string
	// UpdatedAt is the backend object modification time when known.
	UpdatedAt *time.Time
}
