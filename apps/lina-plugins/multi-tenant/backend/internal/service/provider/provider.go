// Package provider implements the multi-tenant plugin provider adapter.
package provider

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"

	pkgtenantcap "lina-core/pkg/tenantcap"
	"lina-plugin-multi-tenant/backend/internal/service/membership"
	"lina-plugin-multi-tenant/backend/internal/service/resolver"
	"lina-plugin-multi-tenant/backend/internal/service/resolverconfig"
)

// Provider is the plugin-owned tenant capability provider. It mirrors the
// host tenantcap contract so registration can be wired once the host seam lands.
type Provider struct {
	membershipSvc     membership.Service
	resolverSvc       resolver.Service
	resolverConfigSvc resolverconfig.Service
}

// Ensure Provider implements the host tenant capability provider contract.
var _ pkgtenantcap.Provider = (*Provider)(nil)
var _ pkgtenantcap.UserMembershipProvider = (*Provider)(nil)

// New creates and returns a Provider instance.
func New(membershipSvc membership.Service, resolverSvc resolver.Service, resolverConfigSvc resolverconfig.Service) *Provider {
	return &Provider{
		membershipSvc:     membershipSvc,
		resolverSvc:       resolverSvc,
		resolverConfigSvc: resolverConfigSvc,
	}
}

// ResolveTenant resolves a tenant from request metadata.
func (p *Provider) ResolveTenant(ctx context.Context, request *ghttp.Request) (*pkgtenantcap.ResolverResult, error) {
	config, err := p.resolverConfigSvc.Get(ctx)
	if err != nil {
		return nil, err
	}
	result, err := p.resolverSvc.Resolve(ctx, request, resolverconfig.ToResolverConfig(config))
	if err != nil {
		return nil, err
	}
	return &pkgtenantcap.ResolverResult{
		TenantID:        pkgtenantcap.TenantID(result.TenantID),
		Matched:         true,
		ActingAsTenant:  result.ActingAsTenant,
		IsImpersonation: result.ActingAsTenant,
	}, nil
}

// ValidateUserInTenant validates one user belongs to one tenant.
func (p *Provider) ValidateUserInTenant(ctx context.Context, userID int, tenantID pkgtenantcap.TenantID) error {
	_, err := p.membershipSvc.GetByUserAndTenant(ctx, int64(userID), int64(tenantID))
	return err
}

// ListUserTenants returns tenant options for one user.
func (p *Provider) ListUserTenants(ctx context.Context, userID int) ([]pkgtenantcap.TenantInfo, error) {
	tenants, err := p.membershipSvc.ListUserTenants(ctx, int64(userID))
	if err != nil {
		return nil, err
	}
	result := make([]pkgtenantcap.TenantInfo, 0, len(tenants))
	for _, item := range tenants {
		if item == nil {
			continue
		}
		result = append(result, pkgtenantcap.TenantInfo{
			ID:     pkgtenantcap.TenantID(item.Id),
			Code:   item.Code,
			Name:   item.Name,
			Status: item.Status,
		})
	}
	return result, nil
}

// SwitchTenant validates one user can switch to a target tenant.
func (p *Provider) SwitchTenant(ctx context.Context, userID int, target pkgtenantcap.TenantID) error {
	return p.ValidateUserInTenant(ctx, userID, target)
}

// ApplyUserTenantScope constrains user rows by active current-tenant membership.
func (p *Provider) ApplyUserTenantScope(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
) (*gdb.Model, bool, error) {
	return p.membershipSvc.ApplyUserTenantScope(ctx, model, userIDColumn)
}

// ApplyUserTenantFilter constrains platform user-list rows to a requested tenant.
func (p *Provider) ApplyUserTenantFilter(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	tenantID pkgtenantcap.TenantID,
) (*gdb.Model, bool, error) {
	return p.membershipSvc.ApplyUserTenantFilter(ctx, model, userIDColumn, tenantID)
}

// ListUserTenantProjections returns tenant ownership labels for visible users.
func (p *Provider) ListUserTenantProjections(
	ctx context.Context,
	userIDs []int,
) (map[int]*pkgtenantcap.UserTenantProjection, error) {
	return p.membershipSvc.ListUserTenantProjections(ctx, userIDs)
}

// ResolveUserTenantAssignment validates requested memberships and returns a host write plan.
func (p *Provider) ResolveUserTenantAssignment(
	ctx context.Context,
	requested []pkgtenantcap.TenantID,
	mode pkgtenantcap.UserTenantAssignmentMode,
) (*pkgtenantcap.UserTenantAssignmentPlan, error) {
	return p.membershipSvc.ResolveUserTenantAssignment(ctx, requested, mode)
}

// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
func (p *Provider) ReplaceUserTenantAssignments(
	ctx context.Context,
	userID int,
	plan *pkgtenantcap.UserTenantAssignmentPlan,
) error {
	return p.membershipSvc.ReplaceUserTenantAssignments(ctx, userID, plan)
}

// EnsureUsersInTenant verifies every user has active membership in the tenant.
func (p *Provider) EnsureUsersInTenant(ctx context.Context, userIDs []int, tenantID pkgtenantcap.TenantID) error {
	return p.membershipSvc.EnsureUsersInTenant(ctx, userIDs, tenantID)
}

// ValidateStartupConsistency returns user-membership startup consistency failures.
func (p *Provider) ValidateStartupConsistency(ctx context.Context) ([]string, error) {
	return p.membershipSvc.ValidateStartupConsistency(ctx)
}
