// This file implements the stream alias detail controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
)

// GetAlias returns one stream alias detail.
func (c *ControllerV1) GetAlias(ctx context.Context, req *v1.GetAliasReq) (res *v1.GetAliasRes, err error) {
	out, err := c.mediaSvc.GetAlias(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetAliasRes{
		Id:         out.Id,
		Alias:      out.Alias,
		AutoRemove: out.AutoRemove,
		StreamPath: out.StreamPath,
		CreateTime: out.CreateTime,
	}, nil
}
