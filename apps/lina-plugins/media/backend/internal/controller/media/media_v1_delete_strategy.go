// This file implements the media strategy deletion controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
)

// DeleteStrategy deletes one media strategy.
func (c *ControllerV1) DeleteStrategy(ctx context.Context, req *v1.DeleteStrategyReq) (res *v1.DeleteStrategyRes, err error) {
	if err = c.mediaSvc.DeleteStrategy(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteStrategyRes{Id: req.Id}, nil
}
