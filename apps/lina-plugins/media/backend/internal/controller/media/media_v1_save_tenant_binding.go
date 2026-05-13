// This file implements the tenant strategy binding save controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// SaveTenantBinding creates or updates one tenant strategy binding.
func (c *ControllerV1) SaveTenantBinding(ctx context.Context, req *v1.SaveTenantBindingReq) (res *v1.SaveTenantBindingRes, err error) {
	out, err := c.mediaSvc.SaveTenantBinding(ctx, mediasvc.TenantBindingMutationInput{
		TenantId:   req.TenantId,
		StrategyId: req.StrategyId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SaveTenantBindingRes{
		TenantId:   out.TenantId,
		StrategyId: out.StrategyId,
	}, nil
}
