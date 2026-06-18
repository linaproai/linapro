// Package filecap defines governed file-domain capability contracts for
// plugins without exposing host file storage tables or physical paths.
package filecap

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"
)

const (
	// MaxSearchPageSize limits one file candidate search page.
	MaxSearchPageSize = 200
)

// FileID identifies one governed file resource.
type FileID string

// FileProjection describes one file reference visible to a plugin.
type FileProjection struct {
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

// SearchInput constrains governed file candidate search.
type SearchInput struct {
	// BusinessScene filters by host file usage scene.
	BusinessScene string
	// Keyword filters by original or stored file name.
	Keyword string
	// MimeType filters by coarse MIME type inferred from file suffix.
	MimeType string
	// Page constrains result size and stable sorting.
	Page capmodel.PageRequest
}

// Service defines read-oriented file capability methods.
type Service interface {
	// BatchGet returns visible file projections and opaque missing IDs.
	BatchGet(ctx context.Context, capCtx capmodel.CapabilityContext, ids []FileID) (*capmodel.BatchResult[*FileProjection, FileID], error)
	// Search returns one bounded page of visible file projections.
	Search(ctx context.Context, capCtx capmodel.CapabilityContext, input SearchInput) (*capmodel.PageResult[*FileProjection], error)
	// EnsureVisible rejects when any requested file is absent or invisible.
	EnsureVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []FileID) error
}

// AdminService defines governed file management commands.
type AdminService interface {
	// Delete deletes visible files after target and scene checks.
	Delete(ctx context.Context, capCtx capmodel.CapabilityContext, ids []FileID) error
}

// ScopeService defines host-internal file visibility helpers.
type ScopeService interface {
	// EnsureVisible rejects when any file is outside caller scope.
	EnsureVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []FileID) error
}
