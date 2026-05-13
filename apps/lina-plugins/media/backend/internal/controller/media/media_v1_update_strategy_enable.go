// This file implements the strategy enable-state update controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
)

// UpdateStrategyEnable changes one media strategy enable status.
func (c *ControllerV1) UpdateStrategyEnable(ctx context.Context, req *v1.UpdateStrategyEnableReq) (res *v1.UpdateStrategyEnableRes, err error) {
	if err = c.mediaSvc.UpdateStrategyEnable(ctx, req.Id, req.Enable); err != nil {
		return nil, err
	}
	return &v1.UpdateStrategyEnableRes{Id: req.Id}, nil
}
