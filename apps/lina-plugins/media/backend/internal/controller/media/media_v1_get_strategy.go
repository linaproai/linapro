// This file implements the media strategy detail controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
)

// GetStrategy returns one media strategy detail.
func (c *ControllerV1) GetStrategy(ctx context.Context, req *v1.GetStrategyReq) (res *v1.GetStrategyRes, err error) {
	out, err := c.mediaSvc.GetStrategy(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetStrategyRes{
		Id:        out.Id,
		Name:      out.Name,
		Strategy:  out.Strategy,
		Global:    out.Global,
		Enable:    out.Enable,
		CreatorId: out.CreatorId,
		UpdaterId: out.UpdaterId,
		CreateTime: out.CreateTime,
		UpdateTime: out.UpdateTime,
	}, nil
}
