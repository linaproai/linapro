// This file publishes the narrow token-bound role access projection consumed by
// dynamic plugin route authentication.

package role

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/datascope"
)

// DynamicRouteAccessProjection is the role-owned permission and data-scope view
// needed by dynamic plugin route authentication and bridge identity snapshots.
type DynamicRouteAccessProjection struct {
	Permissions          []string        // Permissions contains effective permission strings for the token tenant.
	RoleNames            []string        // RoleNames contains enabled role display names in the token tenant.
	DataScope            datascope.Scope // DataScope is the effective role data-scope for governed resources.
	DataScopeUnsupported bool            // DataScopeUnsupported reports unsupported role data-scope values.
	UnsupportedDataScope int             // UnsupportedDataScope stores the first unsupported scope value.
	IsSuperAdmin         bool            // IsSuperAdmin reports whether the user has host super-admin access.
}

// BuildDynamicRouteAccessProjection returns a detached access projection for one
// token, user, and tenant by reusing the token access snapshot owner.
func (s *serviceImpl) BuildDynamicRouteAccessProjection(
	ctx context.Context,
	tokenID string,
	userID int,
	tenantID int,
) (*DynamicRouteAccessProjection, error) {
	tokenID = strings.TrimSpace(tokenID)
	if tokenID == "" {
		return nil, gerror.New("dynamic route access projection token ID is required")
	}
	if userID <= 0 {
		return nil, gerror.New("dynamic route access projection user ID is required")
	}
	if tenantID < 0 {
		return nil, gerror.New("dynamic route access projection tenant ID is invalid")
	}
	accessCtx := dynamicRouteAccessContext(ctx, tokenID, tenantID)
	access, err := s.getTokenAccessContext(accessCtx, tokenID, userID)
	if err != nil {
		return nil, err
	}
	if access == nil {
		return &DynamicRouteAccessProjection{
			Permissions: []string{},
			RoleNames:   []string{},
			DataScope:   datascope.ScopeNone,
		}, nil
	}
	return &DynamicRouteAccessProjection{
		Permissions:          cloneSliceWithCopy(access.Permissions),
		RoleNames:            cloneSliceWithCopy(access.RoleNames),
		DataScope:            access.DataScope,
		DataScopeUnsupported: access.DataScopeUnsupported,
		UnsupportedDataScope: access.UnsupportedDataScope,
		IsSuperAdmin:         access.IsSuperAdmin,
	}, nil
}

// dynamicRouteAccessContext fixes the token tenant and token ID before handing
// projection reads to the shared role access snapshot cache.
func dynamicRouteAccessContext(ctx context.Context, tokenID string, tenantID int) context.Context {
	scoped := datascope.WithTenantScope(ctx, tenantID)
	return context.WithValue(scoped, dynamicRouteAccessTokenContextKey{}, strings.TrimSpace(tokenID))
}

// dynamicRouteAccessTokenContextKey carries the token ID for projection reads
// that originate outside the normal HTTP middleware business context.
type dynamicRouteAccessTokenContextKey struct{}
