// This file defines the source-plugin visible tenant-filter contract.

package contract

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
)

// TenantFilterColumn is the shared tenant column name used by tenant-scoped plugin tables.
const TenantFilterColumn = "tenant_id"

// TenantFilterService defines tenant filtering operations published to source plugins.
type TenantFilterService interface {
	// Context returns the plugin-visible tenant, actor, impersonation, and
	// platform-bypass metadata resolved from the host business context.
	Context(ctx context.Context) TenantFilterContext
	// Apply adds the conventional tenant_id predicate to one model unless the
	// current context permits platform bypass. Qualifier may contain only the
	// table name or alias used by joined queries; an empty qualifier applies the
	// unqualified tenant_id column, and a nil model is returned unchanged.
	Apply(ctx context.Context, model *gdb.Model, qualifier string) *gdb.Model
}

// PlatformBypassEvaluator defines the optional host policy used to decide
// whether a platform request can read across tenant-owned plugin rows.
type PlatformBypassEvaluator interface {
	// PlatformBypass reports whether the current request may bypass tenant filtering.
	PlatformBypass(ctx context.Context) bool
}

// TenantFilterContext carries the plugin-visible tenant and audit identity metadata.
type TenantFilterContext struct {
	// UserID is the authenticated user bound to the current request.
	UserID int
	// TenantID is the current request tenant.
	TenantID int
	// ActingUserID is the real actor to persist in audit records.
	ActingUserID int
	// OnBehalfOfTenantID is set only when the request acts on behalf of a tenant.
	OnBehalfOfTenantID int
	// ActingAsTenant reports whether the request acts through a tenant view.
	ActingAsTenant bool
	// IsImpersonation marks platform impersonation.
	IsImpersonation bool
	// PlatformBypass reports whether the request runs in platform scope.
	PlatformBypass bool
}
