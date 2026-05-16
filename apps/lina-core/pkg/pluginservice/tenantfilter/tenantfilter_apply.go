// This file contains the concrete tenant-context derivation and model filtering
// behavior for source plugins. It keeps query mutation and bypass handling out
// of the package entrypoint while preserving tenant filter contract semantics.

package tenantfilter

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/pluginservice/contract"
)

// Current returns the current tenant ID from the host business context.
func (s *serviceImpl) Current(ctx context.Context) int {
	return s.CurrentContext(ctx).TenantID
}

// CurrentContext returns tenant and impersonation metadata from host business context.
func (s *serviceImpl) CurrentContext(ctx context.Context) contract.TenantFilterContext {
	current := s.bizCtxSvc.Current(ctx)
	actingUserID := current.ActingUserID
	if actingUserID == 0 {
		actingUserID = current.UserID
	}
	onBehalfOfTenantID := 0
	if current.IsImpersonation || current.ActingAsTenant {
		onBehalfOfTenantID = current.TenantID
	}
	platformBypass := current.PlatformBypass
	if s.bypassEvaluator != nil {
		platformBypass = s.bypassEvaluator.PlatformBypass(ctx)
	}
	return contract.TenantFilterContext{
		UserID:             current.UserID,
		TenantID:           current.TenantID,
		ActingUserID:       actingUserID,
		OnBehalfOfTenantID: onBehalfOfTenantID,
		ActingAsTenant:     current.ActingAsTenant,
		IsImpersonation:    current.IsImpersonation,
		PlatformBypass:     platformBypass,
	}
}

// Apply adds tenant filtering to one model using the conventional tenant column.
func (s *serviceImpl) Apply(ctx context.Context, model *gdb.Model) *gdb.Model {
	return s.ApplyColumn(ctx, model, contract.TenantFilterColumn)
}

// ApplyColumn adds tenant filtering with an explicit column for joined queries.
func (s *serviceImpl) ApplyColumn(ctx context.Context, model *gdb.Model, column string) *gdb.Model {
	if model == nil {
		return nil
	}
	current := s.CurrentContext(ctx)
	if current.PlatformBypass {
		return model
	}
	tenantColumn := strings.TrimSpace(column)
	if tenantColumn == "" {
		tenantColumn = contract.TenantFilterColumn
	}
	return model.Where(tenantColumn, current.TenantID)
}
