// Package authz defines authorization-domain capability contracts for plugins
// without exposing host role, menu, or permission tables.
package authz

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"
)

// Service defines governed authorization capability methods. Reads reuse the
// caller permission snapshot or bounded info queries; mutations must
// validate assignable permissions, tenant/resource scope, audit source, and
// transaction-after cache revision impact.
type Service interface {
	// BatchGetPermissions returns visible permission info records and opaque missing keys.
	BatchGetPermissions(ctx context.Context, keys []PermissionKey) (*capmodel.BatchResult[*PermissionInfo, PermissionKey], error)
	// BatchHasPermissions reports whether the actor has each permission key in the current scope.
	BatchHasPermissions(ctx context.Context, keys []PermissionKey) (map[PermissionKey]bool, error)
	// HasPermission reports whether the actor has one permission in the current scope.
	HasPermission(ctx context.Context, key PermissionKey) (bool, error)
	// IsPlatformAdmin reports whether the user has a platform all-data role.
	IsPlatformAdmin(ctx context.Context, userID UserID) (bool, error)
	// ReplaceRolePermissions replaces one role's permission set after target,
	// permission-key, tenant, data-scope, transaction, audit, and permission
	// cache-revision checks.
	ReplaceRolePermissions(ctx context.Context, roleID RoleID, keys []PermissionKey) error
}

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

// PermissionInfo describes one permission visible to a plugin.
type PermissionInfo struct {
	// Key is the stable permission key.
	Key PermissionKey
	// LabelKey is the stable runtime i18n label key.
	LabelKey string
	// Label is the optional locale-resolved label.
	Label string
}
