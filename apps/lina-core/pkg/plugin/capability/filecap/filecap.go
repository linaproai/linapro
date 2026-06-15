// Package filecap defines governed file-domain capability contracts for
// plugins without exposing host file storage tables or physical paths.
package filecap

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"
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

// Service defines read-oriented file capability methods.
type Service interface {
	// BatchGet returns visible file projections and opaque missing IDs.
	BatchGet(ctx context.Context, capCtx capmodel.CapabilityContext, ids []FileID) (*capmodel.BatchResult[*FileProjection, FileID], error)
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
