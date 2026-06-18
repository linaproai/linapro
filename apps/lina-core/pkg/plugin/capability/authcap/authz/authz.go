// Package authz defines authorization-domain capability contracts for plugins
// without exposing host role, menu, or permission tables.
package authz

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"
)

const (
	// MaxBatchHasPermissions limits one authorization batch-boolean request.
	MaxBatchHasPermissions = 200
)

// RoleID identifies one role in plugin-visible domain boundaries.
type RoleID string

// PermissionKey identifies one governed permission string.
type PermissionKey string

// UserID identifies a user for authorization-domain checks.
type UserID string

// PermissionProjection describes one permission visible to a plugin.
type PermissionProjection struct {
	// Key is the stable permission key.
	Key PermissionKey
	// LabelKey is the stable runtime i18n label key.
	LabelKey string
	// Label is the optional locale-resolved label.
	Label string
}

// Service defines read-oriented authorization capability methods.
type Service interface {
	// BatchGetPermissions returns visible permission projections and opaque missing keys.
	BatchGetPermissions(ctx context.Context, capCtx capmodel.CapabilityContext, keys []PermissionKey) (*capmodel.BatchResult[*PermissionProjection, PermissionKey], error)
	// BatchHasPermissions reports whether the actor has each permission key in the current scope.
	BatchHasPermissions(ctx context.Context, capCtx capmodel.CapabilityContext, keys []PermissionKey) (map[PermissionKey]bool, error)
	// HasPermission reports whether the actor has one permission in the current scope.
	HasPermission(ctx context.Context, capCtx capmodel.CapabilityContext, key PermissionKey) (bool, error)
	// IsPlatformAdmin reports whether the user has a platform all-data role.
	IsPlatformAdmin(ctx context.Context, capCtx capmodel.CapabilityContext, userID UserID) (bool, error)
}

// AdminService defines authorization management commands.
type AdminService interface {
	// ReplaceRolePermissions replaces one role's permission set after target checks.
	ReplaceRolePermissions(ctx context.Context, capCtx capmodel.CapabilityContext, roleID RoleID, keys []PermissionKey) error
}

// ScopeService defines host-internal authorization helpers.
type ScopeService interface {
	// EnsurePermissionsVisible rejects when any permission is outside caller scope.
	EnsurePermissionsVisible(ctx context.Context, capCtx capmodel.CapabilityContext, keys []PermissionKey) error
}
