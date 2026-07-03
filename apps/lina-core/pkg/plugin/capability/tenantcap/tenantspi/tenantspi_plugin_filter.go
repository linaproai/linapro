// This file defines tenant-filter context and table helpers for same-process
// callers whose plugin-owned tables use the conventional tenant_id discriminator.

package tenantspi

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// TenantFilterColumn is the shared tenant column name used by tenant-scoped plugin tables.
const TenantFilterColumn = "tenant_id"

// TenantFilterContext is the source-plugin alias for plugin-visible tenant filter context.
type TenantFilterContext = tenantcap.TenantFilterContext

// pluginTableFilterService implements the tenant filter context service.
type pluginTableFilterService struct {
	bizCtxSvc bizctxcap.Service
}

// NewPluginTableFilter creates tenant filter context reads from host-owned context dependencies.
func NewPluginTableFilter(bizCtxSvc bizctxcap.Service) (tenantcap.FilterService, error) {
	if bizCtxSvc == nil {
		return nil, gerror.New("tenantfilter requires host bizctx service")
	}
	return &pluginTableFilterService{
		bizCtxSvc: bizCtxSvc,
	}, nil
}

// Context returns tenant and impersonation metadata from host business context.
func (s *pluginTableFilterService) Context(ctx context.Context) TenantFilterContext {
	return tenantFilterContextFromCurrent(s.bizCtxSvc.Current(ctx))
}

// tenantFilterContextFromCurrent converts host business context into the
// serializable tenant filter context shared by source and dynamic plugins.
func tenantFilterContextFromCurrent(current bizctxcap.CurrentContext) TenantFilterContext {
	actingUserID := current.ActingUserID
	if actingUserID == 0 {
		actingUserID = current.UserID
	}
	onBehalfOfTenantID := 0
	if current.IsImpersonation || current.ActingAsTenant {
		onBehalfOfTenantID = current.TenantID
	}
	return TenantFilterContext{
		UserID:             current.UserID,
		TenantID:           current.TenantID,
		ActingUserID:       actingUserID,
		OnBehalfOfTenantID: onBehalfOfTenantID,
		ActingAsTenant:     current.ActingAsTenant,
		IsImpersonation:    current.IsImpersonation,
		PlatformBypass:     current.PlatformBypass,
	}
}

// ApplyPluginTableFilter adds tenant filtering to one model using an optional
// table qualifier. It is intentionally kept in tenantspi so ordinary
// tenantcap.FilterService remains a serializable context contract.
func ApplyPluginTableFilter(ctx context.Context, filter tenantcap.FilterService, model *gdb.Model, qualifier string) *gdb.Model {
	if model == nil {
		return nil
	}
	if filter == nil {
		return model
	}
	current := filter.Context(ctx)
	if current.PlatformBypass {
		return model
	}
	tenantColumn := TenantFilterColumn
	if tenantQualifier := strings.TrimSpace(qualifier); tenantQualifier != "" {
		tenantColumn = fmt.Sprintf("%s.%s", tenantQualifier, TenantFilterColumn)
	}
	return model.Where(tenantColumn, current.TenantID)
}
