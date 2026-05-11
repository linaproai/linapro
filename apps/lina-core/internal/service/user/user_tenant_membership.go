// This file bridges host user management to the tenant capability provider.

package user

import (
	"context"

	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

const (
	// tenantPrimaryFieldUpdateCreate marks a tenant assignment plan for user creation.
	tenantPrimaryFieldUpdateCreate = true
)

// currentTenantID returns the tenant identity currently bound to the request.
func currentTenantID(ctx context.Context) int {
	return datascope.CurrentTenantID(ctx)
}

// platformTenantBypass reports whether the current request can administer
// cross-tenant membership ownership from the platform context.
func (s *serviceImpl) platformTenantBypass(ctx context.Context) bool {
	if s == nil {
		return true
	}
	if s.bizCtxSvc != nil {
		if bizCtx := s.bizCtxSvc.Get(ctx); bizCtx != nil &&
			bizCtx.TenantId == int(pkgtenantcap.PlatformTenantID) &&
			bizCtx.DataScope == int(datascope.ScopeAll) &&
			!bizCtx.DataScopeUnsupported &&
			!bizCtx.ActingAsTenant &&
			!bizCtx.IsImpersonation {
			return true
		}
	}
	if s.tenantSvc != nil && s.tenantSvc.PlatformBypass(ctx) {
		return true
	}
	scope, err := s.currentScopeSvc().Current(ctx)
	if err != nil {
		return false
	}
	return scope != nil &&
		scope.Scope == datascope.ScopeAll &&
		currentTenantID(ctx) == int(pkgtenantcap.PlatformTenantID)
}

// resolveCreateTenantMemberships resolves request tenant ownership for new
// users without letting tenant-scoped requests choose another tenant.
func (s *serviceImpl) resolveCreateTenantMemberships(
	ctx context.Context,
	requestedTenantIDs []int,
) (*pkgtenantcap.UserTenantAssignmentPlan, error) {
	tenantID := currentTenantID(ctx)
	if !s.multiTenantEnabled(ctx) {
		if tenantID == datascope.PlatformTenantID {
			return &pkgtenantcap.UserTenantAssignmentPlan{
				PrimaryTenant: pkgtenantcap.PLATFORM,
			}, nil
		}
		return &pkgtenantcap.UserTenantAssignmentPlan{
			TenantIDs:     []pkgtenantcap.TenantID{pkgtenantcap.TenantID(tenantID)},
			ShouldReplace: tenantPrimaryFieldUpdateCreate,
			PrimaryTenant: pkgtenantcap.TenantID(tenantID),
		}, nil
	}
	normalized := normalizeTenantIDs(requestedTenantIDs)
	if tenantID > datascope.PlatformTenantID {
		if err := ensureRequestedTenantsMatchCurrentTenant(tenantID, normalized); err != nil {
			return nil, err
		}
	} else {
		if err := s.ensurePlatformRequestedTenantsAllowed(ctx, normalized); err != nil {
			return nil, err
		}
	}
	return s.tenantSvc.ResolveUserTenantAssignment(
		ctx,
		toTenantIDs(requestedTenantIDs),
		pkgtenantcap.UserTenantAssignmentCreate,
	)
}

// resolveUpdateTenantMemberships resolves whether user membership should be
// replaced and which tenant IDs should be stored.
func (s *serviceImpl) resolveUpdateTenantMemberships(
	ctx context.Context,
	requestedTenantIDs []int,
) (*pkgtenantcap.UserTenantAssignmentPlan, error) {
	if !s.multiTenantEnabled(ctx) || requestedTenantIDs == nil {
		return &pkgtenantcap.UserTenantAssignmentPlan{}, nil
	}
	normalized := normalizeTenantIDs(requestedTenantIDs)
	if tenantID := currentTenantID(ctx); tenantID > datascope.PlatformTenantID {
		if err := ensureRequestedTenantsMatchCurrentTenant(tenantID, normalized); err != nil {
			return nil, err
		}
	} else {
		if err := s.ensurePlatformRequestedTenantsAllowed(ctx, normalized); err != nil {
			return nil, err
		}
	}
	return s.tenantSvc.ResolveUserTenantAssignment(
		ctx,
		toTenantIDs(requestedTenantIDs),
		pkgtenantcap.UserTenantAssignmentUpdate,
	)
}

// ensurePlatformRequestedTenantsAllowed rejects platform-context membership
// writes outside the current operator's active tenant membership set.
func (s *serviceImpl) ensurePlatformRequestedTenantsAllowed(
	ctx context.Context,
	normalized []int,
) error {
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx == nil || bizCtx.UserId <= 0 {
		if len(normalized) == 0 {
			return nil
		}
		return bizerr.NewCode(pkgtenantcap.CodeTenantForbidden, bizerr.P("tenantId", normalized[0]))
	}
	tenants, err := s.tenantSvc.ListUserTenants(ctx, bizCtx.UserId)
	if err != nil {
		return err
	}
	if len(tenants) == 0 {
		if len(normalized) > 0 && !s.platformTenantBypass(ctx) {
			return bizerr.NewCode(pkgtenantcap.CodePlatformPermissionRequired)
		}
		return nil
	}
	if len(normalized) == 0 {
		return bizerr.NewCode(pkgtenantcap.CodeCrossTenantNotAllowed)
	}
	owned := make(map[int]struct{}, len(tenants))
	for _, tenant := range tenants {
		if tenant.ID > pkgtenantcap.PLATFORM {
			owned[int(tenant.ID)] = struct{}{}
		}
	}
	for _, tenantID := range normalized {
		if _, ok := owned[tenantID]; !ok {
			return bizerr.NewCode(pkgtenantcap.CodeTenantForbidden, bizerr.P("tenantId", tenantID))
		}
	}
	return nil
}

// ensureListTenantFilterAllowed rejects platform-context user-list filters for
// tenants outside the current operator's active membership set.
func (s *serviceImpl) ensureListTenantFilterAllowed(ctx context.Context, tenantID int) error {
	normalized := normalizeTenantIDs([]int{tenantID})
	if len(normalized) == 0 || currentTenantID(ctx) != datascope.PlatformTenantID {
		return nil
	}
	return s.ensurePlatformRequestedTenantsAllowed(ctx, normalized)
}

// ensureRequestedTenantsMatchCurrentTenant rejects tenant-context requests that
// try to write another tenant's membership ownership.
func ensureRequestedTenantsMatchCurrentTenant(currentTenantID int, normalized []int) error {
	for _, tenantID := range normalized {
		if tenantID != currentTenantID {
			return bizerr.NewCode(pkgtenantcap.CodeCrossTenantNotAllowed)
		}
	}
	return nil
}

// normalizeTenantIDs returns positive unique tenant IDs while preserving order.
func normalizeTenantIDs(tenantIDs []int) []int {
	normalized := make([]int, 0, len(tenantIDs))
	seen := make(map[int]struct{}, len(tenantIDs))
	for _, tenantID := range tenantIDs {
		if tenantID <= datascope.PlatformTenantID {
			continue
		}
		if _, ok := seen[tenantID]; ok {
			continue
		}
		seen[tenantID] = struct{}{}
		normalized = append(normalized, tenantID)
	}
	return normalized
}

// toTenantIDs converts host request tenant IDs to the capability contract type.
func toTenantIDs(tenantIDs []int) []pkgtenantcap.TenantID {
	normalized := normalizeTenantIDs(tenantIDs)
	result := make([]pkgtenantcap.TenantID, 0, len(normalized))
	for _, tenantID := range normalized {
		result = append(result, pkgtenantcap.TenantID(tenantID))
	}
	return result
}

// multiTenantEnabled reports whether the optional multi-tenant plugin is active.
func (s *serviceImpl) multiTenantEnabled(ctx context.Context) bool {
	return s != nil && s.tenantSvc != nil && s.tenantSvc.Enabled(ctx)
}

// GetUserTenantMemberships returns tenant ownership data for one visible user.
func (s *serviceImpl) GetUserTenantMemberships(ctx context.Context, userId int) ([]int, []string, error) {
	if !s.multiTenantEnabled(ctx) {
		return []int{}, []string{}, nil
	}
	if err := s.ensureUserVisible(ctx, userId); err != nil {
		return nil, nil, err
	}
	items, err := s.tenantSvc.ListUserTenantProjections(ctx, []int{userId})
	if err != nil {
		return nil, nil, err
	}
	item := items[userId]
	if item == nil {
		return []int{}, []string{}, nil
	}
	tenantIDs := make([]int, 0, len(item.TenantIDs))
	for _, tenantID := range item.TenantIDs {
		tenantIDs = append(tenantIDs, int(tenantID))
	}
	return tenantIDs, item.TenantNames, nil
}
