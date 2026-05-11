// Package tenantfilter exposes shared tenant query helpers for source plugins
// whose plugin-owned tables use the conventional tenant_id discriminator.
package tenantfilter

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/pluginservice/bizctx"
)

// Column is the shared tenant column name used by tenant-scoped plugin tables.
const Column = "tenant_id"

// Context carries the plugin-visible tenant and audit identity metadata.
type Context struct {
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

// Current returns the current tenant ID from the host business context.
func Current(ctx context.Context) int {
	return CurrentContext(ctx).TenantID
}

// CurrentContext returns tenant and impersonation metadata from host business context.
func CurrentContext(ctx context.Context) Context {
	current := bizctx.New().Current(ctx)
	actingUserID := current.ActingUserID
	if actingUserID == 0 {
		actingUserID = current.UserID
	}
	onBehalfOfTenantID := 0
	if current.IsImpersonation || current.ActingAsTenant {
		onBehalfOfTenantID = current.TenantID
	}
	return Context{
		UserID:             current.UserID,
		TenantID:           current.TenantID,
		ActingUserID:       actingUserID,
		OnBehalfOfTenantID: onBehalfOfTenantID,
		ActingAsTenant:     current.ActingAsTenant,
		IsImpersonation:    current.IsImpersonation,
		PlatformBypass:     current.PlatformBypass,
	}
}

// Apply adds tenant filtering to one model using the conventional tenant column.
func Apply(ctx context.Context, model *gdb.Model) *gdb.Model {
	return ApplyColumn(ctx, model, Column)
}

// ApplyColumn adds tenant filtering with an explicit column for joined queries.
func ApplyColumn(ctx context.Context, model *gdb.Model, column string) *gdb.Model {
	if model == nil {
		return nil
	}
	tenantColumn := strings.TrimSpace(column)
	if tenantColumn == "" {
		tenantColumn = Column
	}
	return model.Where(tenantColumn, Current(ctx))
}
