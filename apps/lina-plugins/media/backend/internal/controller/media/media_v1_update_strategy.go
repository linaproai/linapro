// This file implements the media strategy update controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// UpdateStrategy updates one media strategy.
func (c *ControllerV1) UpdateStrategy(ctx context.Context, req *v1.UpdateStrategyReq) (res *v1.UpdateStrategyRes, err error) {
	if err = c.mediaSvc.UpdateStrategy(ctx, req.Id, mediasvc.StrategyMutationInput{
		Name:     req.Name,
		Strategy: req.Strategy,
		Enable:   req.Enable,
		Global:   req.Global,
	}); err != nil {
		return nil, err
	}
	return &v1.UpdateStrategyRes{Id: req.Id}, nil
}
