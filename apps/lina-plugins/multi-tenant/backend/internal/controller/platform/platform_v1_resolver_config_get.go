// This file implements the platform resolver-policy detail endpoint.

package platform

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/platform/v1"
)

// ResolverConfigGet returns built-in resolver policy.
func (c *ControllerV1) ResolverConfigGet(ctx context.Context, _ *v1.ResolverConfigGetReq) (res *v1.ResolverConfigGetRes, err error) {
	config, err := c.resolverConfigSvc.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ResolverConfigGetRes{ResolverConfigEntity: toAPIResolverConfig(config)}, nil
}
