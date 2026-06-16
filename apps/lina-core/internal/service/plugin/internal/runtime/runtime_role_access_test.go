// This file provides role-access projector fakes for runtime package tests.

package runtime

import (
	"context"

	"lina-core/internal/service/datascope"
	rolesvc "lina-core/internal/service/role"
)

// testRoleAccessProjector returns deterministic access projections for tests
// that only need constructor wiring or route-auth behavior.
type testRoleAccessProjector struct {
	projection *rolesvc.DynamicRouteAccessProjection
	err        error
}

// BuildDynamicRouteAccessProjection returns the configured projection.
func (p testRoleAccessProjector) BuildDynamicRouteAccessProjection(
	context.Context,
	string,
	int,
	int,
) (*rolesvc.DynamicRouteAccessProjection, error) {
	if p.err != nil {
		return nil, p.err
	}
	if p.projection == nil {
		return &rolesvc.DynamicRouteAccessProjection{
			Permissions: []string{},
			RoleNames:   []string{},
			DataScope:   datascope.ScopeNone,
		}, nil
	}
	return &rolesvc.DynamicRouteAccessProjection{
		Permissions:          append([]string(nil), p.projection.Permissions...),
		RoleNames:            append([]string(nil), p.projection.RoleNames...),
		DataScope:            p.projection.DataScope,
		DataScopeUnsupported: p.projection.DataScopeUnsupported,
		UnsupportedDataScope: p.projection.UnsupportedDataScope,
		IsSuperAdmin:         p.projection.IsSuperAdmin,
	}, nil
}
