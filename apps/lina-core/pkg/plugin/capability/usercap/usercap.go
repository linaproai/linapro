// Package usercap defines the stable user-domain capability contract exposed to
// plugins without leaking sys_user storage or host DAO models.
package usercap

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"
)

// UserID identifies a user at plugin-visible domain boundaries.
type UserID string

// UserProjection is the minimal user display projection exposed to plugins.
type UserProjection struct {
	// ID is the user domain identifier.
	ID UserID
	// Username is the stable login name.
	Username string
	// Nickname is the display name.
	Nickname string
	// Avatar is an optional avatar URL or governed file reference.
	Avatar string
	// Status is the stable user lifecycle status.
	Status string
	// TenantID is the owning tenant domain identifier.
	TenantID capmodel.DomainID
	// LabelKey is the optional i18n key for synthetic labels.
	LabelKey string
	// Label is the optional locale-resolved display label.
	Label string
}

// SearchInput constrains user candidate searches.
type SearchInput struct {
	// Keyword filters visible users by username, nickname, or phone/email owner fields.
	Keyword string
	// Page constrains candidate size and sorting.
	Page capmodel.PageRequest
}

// Service defines read-oriented user capability methods for plugins.
type Service interface {
	// BatchGet returns visible user projections and opaque missing IDs.
	BatchGet(ctx context.Context, capCtx capmodel.CapabilityContext, ids []UserID) (*capmodel.BatchResult[*UserProjection, UserID], error)
	// Search searches visible user candidates with bounded paging.
	Search(ctx context.Context, capCtx capmodel.CapabilityContext, input SearchInput) (*capmodel.PageResult[*UserProjection], error)
	// EnsureVisible rejects when any requested user is absent or invisible.
	EnsureVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []UserID) error
}

// AdminService defines user management commands exposed through governed
// source-plugin AdminServices or authorized dynamic host service methods.
type AdminService interface {
	// SetStatus changes one user's lifecycle status after target visibility checks.
	SetStatus(ctx context.Context, capCtx capmodel.CapabilityContext, id UserID, status string) error
}

// ScopeService defines host-internal user scope helpers and must not be exposed
// through ordinary plugin service directories.
type ScopeService interface {
	// EnsureVisible rejects when any user is outside the caller data scope.
	EnsureVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []UserID) error
}
