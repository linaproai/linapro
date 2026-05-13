// This file implements the global strategy selection controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
)

// SetGlobalStrategy sets one media strategy as global.
func (c *ControllerV1) SetGlobalStrategy(ctx context.Context, req *v1.SetGlobalStrategyReq) (res *v1.SetGlobalStrategyRes, err error) {
	if err = c.mediaSvc.SetGlobalStrategy(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.SetGlobalStrategyRes{Id: req.Id}, nil
}
