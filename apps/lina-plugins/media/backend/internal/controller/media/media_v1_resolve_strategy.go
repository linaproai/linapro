// This file implements the effective strategy resolution controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ResolveStrategy resolves the effective strategy for a tenant and device pair.
func (c *ControllerV1) ResolveStrategy(ctx context.Context, req *v1.ResolveStrategyReq) (res *v1.ResolveStrategyRes, err error) {
	out, err := c.mediaSvc.ResolveStrategy(ctx, mediasvc.ResolveStrategyInput{
		TenantId: req.TenantId,
		DeviceId: req.DeviceId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ResolveStrategyRes{
		Matched:      out.Matched,
		Source:       out.Source,
		SourceLabel:  out.SourceLabel,
		StrategyId:   out.StrategyId,
		StrategyName: out.StrategyName,
		Strategy:     out.Strategy,
	}, nil
}
