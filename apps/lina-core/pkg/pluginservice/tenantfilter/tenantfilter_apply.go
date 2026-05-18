// This file contains the concrete tenant-context derivation and model filtering
// behavior for source plugins. It keeps query mutation and bypass handling out
// of the package entrypoint while preserving tenant filter contract semantics.

package tenantfilter

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/pluginservice/contract"
)

// Context returns tenant and impersonation metadata from host business context.
func (s *serviceImpl) Context(ctx context.Context) contract.TenantFilterContext {
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

// Apply adds tenant filtering to one model using an optional table qualifier.
func (s *serviceImpl) Apply(ctx context.Context, model *gdb.Model, qualifier string) *gdb.Model {
	if model == nil {
		return nil
	}
	current := s.Context(ctx)
	if current.PlatformBypass {
		return model
	}
	tenantColumn := contract.TenantFilterColumn
	if tenantQualifier := strings.TrimSpace(qualifier); tenantQualifier != "" {
		tenantColumn = fmt.Sprintf("%s.%s", tenantQualifier, contract.TenantFilterColumn)
	}
	return model.Where(tenantColumn, current.TenantID)
}
