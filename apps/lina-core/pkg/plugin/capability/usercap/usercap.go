// Package usercap defines the stable user-domain capability contract exposed to
// plugins without leaking sys_user storage or host DAO models.
package usercap

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"
)

// Status identifies a plugin-visible user lifecycle state.
type Status string

const (
	// MaxBatchResolveIDs limits one user batch-resolve call by user ID count.
	MaxBatchResolveIDs = 100
	// MaxBatchResolveUsernames limits one user batch-resolve call by username count.
	MaxBatchResolveUsernames = 100
	// MaxBatchResolveContacts limits one user batch-resolve call by phone or email count.
	MaxBatchResolveContacts = 100
	// MaxBatchResolveKeys limits the normalized key count across all resolve dimensions.
	MaxBatchResolveKeys = 300

	// StatusDisabled identifies disabled users.
	StatusDisabled Status = "0"
	// StatusEnabled identifies enabled users.
	StatusEnabled Status = "1"
)

// UserID identifies a user at plugin-visible domain boundaries.
type UserID string

// ResolveKey identifies one requested user lookup key without exposing which
// lookup dimension matched a visible user.
type ResolveKey string

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
	// Status filters by user lifecycle state. Empty includes all visible states.
	Status Status
	// TenantID optionally narrows results to one visible tenant.
	TenantID capmodel.DomainID
	// EnabledOnly is a convenience filter for enabled user candidates.
	EnabledOnly bool
	// Page constrains candidate size and sorting.
	Page capmodel.PageRequest
}

// BatchResolveInput constrains user lookup by stable domain IDs and login or
// contact identifiers. Missing results must not distinguish absent users from
// tenant or data-permission filtering.
type BatchResolveInput struct {
	// IDs contains user domain identifiers.
	IDs []UserID
	// Usernames contains stable login names.
	Usernames []string
	// Contacts contains email addresses or phone numbers.
	Contacts []string
}

// Service defines read-oriented user capability methods for plugins.
type Service interface {
	// Current returns the current actor's visible user projection.
	Current(ctx context.Context, capCtx capmodel.CapabilityContext) (*UserProjection, error)
	// BatchGet returns visible user projections and opaque missing IDs.
	BatchGet(ctx context.Context, capCtx capmodel.CapabilityContext, ids []UserID) (*capmodel.BatchResult[*UserProjection, UserID], error)
	// BatchResolve resolves visible users by IDs, usernames, email addresses, or phone numbers.
	BatchResolve(ctx context.Context, capCtx capmodel.CapabilityContext, input BatchResolveInput) (*capmodel.BatchResult[*UserProjection, ResolveKey], error)
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
