// This file implements the platform resolver-policy validation endpoint.

package platform

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/platform/v1"
	"lina-plugin-multi-tenant/backend/internal/service/resolverconfig"
)

// ResolverConfigUpdate validates resolver policy without mutating runtime state.
func (c *ControllerV1) ResolverConfigUpdate(ctx context.Context, req *v1.ResolverConfigUpdateReq) (res *v1.ResolverConfigUpdateRes, err error) {
	err = c.resolverConfigSvc.Update(ctx, resolverconfig.UpdateInput{
		Chain:              req.Chain,
		ReservedSubdomains: req.ReservedSubdomains,
		RootDomain:         req.RootDomain,
		OnAmbiguous:        req.OnAmbiguous,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ResolverConfigUpdateRes{}, nil
}
