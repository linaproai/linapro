// This file defines the source-plugin visible manifest resource contract.

package manifestcap

import "context"

const (
	// MaxBatchPaths bounds one manifest multi-resource request.
	MaxBatchPaths = 100
	// DefaultListLimit bounds manifest list calls when no limit is provided.
	DefaultListLimit = 100
	// MaxListLimit bounds one manifest list response.
	MaxListLimit = 500
	// MaxResourceBytes bounds one returned manifest resource body.
	MaxResourceBytes = 1 * 1024 * 1024
	// MaxTotalBytes bounds total resource bytes in one multi-resource response.
	MaxTotalBytes = 4 * 1024 * 1024
)

// ServiceFactory creates plugin-scoped manifest resource service views.
type ServiceFactory interface {
	// ForPlugin returns a manifest resource service scoped to pluginID. Blank
	// plugin IDs return a service that rejects reads.
	ForPlugin(pluginID string) Service
	// WithArtifactResources returns a new factory view that can use release-bound
	// artifact resources for pluginID. Paths are relative to manifest/.
	WithArtifactResources(pluginID string, resources map[string][]byte) ServiceFactory
}

// Service defines read-only access to one plugin's manifest resources.
type Service interface {
	// Get returns one raw resource under the current plugin manifest
	// root. Paths are slash-separated and relative to manifest/.
	Get(ctx context.Context, path string) ([]byte, error)
	// GetMany returns raw resources for an explicit bounded path set. Missing
	// resources are reported opaquely in MissingPaths.
	GetMany(ctx context.Context, input GetManyInput) (*GetManyOutput, error)
	// List returns metadata for resources under one bounded manifest prefix.
	List(ctx context.Context, input ListInput) (*ListOutput, error)
	// Exists reports whether one allowed manifest resource exists under the
	// current plugin manifest root.
	Exists(ctx context.Context, path string) (bool, error)
	// Scan unmarshals the selected YAML resource, or the nested key inside it,
	// into target. Missing resources leave target unchanged.
	Scan(ctx context.Context, path string, key string, target any) error
}

// Resource describes one plugin-visible manifest resource snapshot.
type Resource struct {
	// Path is the manifest-relative resource path.
	Path string `json:"path"`
	// Size is the resource byte size when known.
	Size int64 `json:"size"`
}

// ResourceContent describes one manifest resource content snapshot.
type ResourceContent struct {
	// Path is the manifest-relative resource path.
	Path string `json:"path"`
	// Body is the raw resource content.
	Body []byte `json:"body,omitempty"`
}

// GetManyInput carries one bounded multi-resource read request.
type GetManyInput struct {
	// Paths are manifest-relative resource paths.
	Paths []string `json:"paths"`
}

// GetManyOutput carries one bounded multi-resource read response.
type GetManyOutput struct {
	// Resources contains found resource contents.
	Resources []*ResourceContent `json:"resources"`
	// MissingPaths contains requested paths with no returned resource.
	MissingPaths []string `json:"missingPaths,omitempty"`
}

// ListInput carries one bounded manifest list request.
type ListInput struct {
	// Prefix limits resources to a manifest-relative path prefix.
	Prefix string `json:"prefix"`
	// Limit bounds returned resource metadata.
	Limit int `json:"limit,omitempty"`
}

// ListOutput carries one bounded manifest list response.
type ListOutput struct {
	// Resources contains metadata for matching resources.
	Resources []*Resource `json:"resources"`
	// Limit is the effective limit applied by the service.
	Limit int `json:"limit"`
}
