// This file implements the media strategy creation controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// CreateStrategy creates one media strategy.
func (c *ControllerV1) CreateStrategy(ctx context.Context, req *v1.CreateStrategyReq) (res *v1.CreateStrategyRes, err error) {
	id, err := c.mediaSvc.CreateStrategy(ctx, mediasvc.StrategyMutationInput{
		Name:     req.Name,
		Strategy: req.Strategy,
		Enable:   req.Enable,
		Global:   req.Global,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateStrategyRes{Id: id}, nil
}
