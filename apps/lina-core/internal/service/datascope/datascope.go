// Package datascope implements shared role data-permission resolution and
// database-scope injection for host-owned resources.
package datascope

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/orgcap"
)

// Scope represents the effective data range stored on enabled roles.
type Scope int

// Role data-scope values follow sys_role.data_scope.
const (
	// ScopeNone denies governed resource access.
	ScopeNone Scope = 0
	// ScopeAll grants access to all governed rows across tenant boundaries.
	ScopeAll Scope = 1
	// ScopeTenant grants access to all governed rows in the current tenant.
	ScopeTenant Scope = 2
	// ScopeDept grants access to rows owned by users in the current department scope.
	ScopeDept Scope = 3
	// ScopeSelf grants access only to rows owned by the current user.
	ScopeSelf Scope = 4
)

// AccessSnapshot stores the effective role-governed data scope for one user.
type AccessSnapshot struct {
	UserID       int   // UserID owns the effective data-scope snapshot.
	Scope        Scope // Scope is the widest enabled role data-scope for the user.
	IsSuperAdmin bool  // IsSuperAdmin reports whether the user bypasses role data-scope checks.
}

// AccessProvider is the narrow role dependency needed to resolve cached data scopes.
type AccessProvider interface {
	// GetUserDataScopeSnapshot returns the user's effective role data-scope snapshot.
	GetUserDataScopeSnapshot(ctx context.Context, userID int) (*AccessSnapshot, error)
}

// Context stores the resolved data-permission snapshot for one request.
type Context struct {
	UserID       int   // UserID is the authenticated operator user ID.
	Scope        Scope // Scope is the widest effective data-scope.
	IsSuperAdmin bool  // IsSuperAdmin reports whether the user bypasses data scope.
}

// Service defines shared role data-scope operations used by host modules before
// reading, mutating, or aggregating governed user-owned rows.
type Service interface {
	// Current resolves the current request user's effective data-scope snapshot
	// from bizctx and the role access provider. Missing authentication returns a
	// bizerr not-authenticated code; missing role state resolves to ScopeNone.
	Current(ctx context.Context) (*Context, error)
	// ApplyUserScope constrains a model by a user-owner column and returns empty
	// when the caller should short-circuit to no rows. Department scope delegates
	// to orgcap when available and otherwise falls back to self scope.
	ApplyUserScope(ctx context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error)
	// ApplyUserScopeWithBypass constrains a model by a user-owner column while
	// preserving rows that match an explicit bypass condition, such as system
	// owned jobs. The bypass branch is composed at database-query time so rows
	// outside data scope are not fetched and filtered in memory.
	ApplyUserScopeWithBypass(ctx context.Context, model *gdb.Model, userIDColumn string, bypassColumn string, bypassValue any) (*gdb.Model, bool, error)
	// EnsureUsersVisible verifies all target user IDs are visible under the
	// current role data scope before write or relationship operations proceed.
	EnsureUsersVisible(ctx context.Context, userIDs []int) error
	// EnsureRowsVisible verifies the caller-provided row set remains visible
	// after scope injection; mismatches return a data-scope denied bizerr.
	EnsureRowsVisible(ctx context.Context, model *gdb.Model, userIDColumn string, expectedCount int) error
}

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc bizctx.Service
	roleSvc   AccessProvider
	orgCapSvc orgcap.Service
}

// New creates one shared data-scope service.
func New(bizCtxSvc bizctx.Service, roleSvc AccessProvider, orgCapSvc orgcap.Service) Service {
	return &serviceImpl{
		bizCtxSvc: bizCtxSvc,
		roleSvc:   roleSvc,
		orgCapSvc: orgCapSvc,
	}
}
