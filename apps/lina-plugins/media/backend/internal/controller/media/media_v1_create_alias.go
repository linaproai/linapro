// This file implements the stream alias creation controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// CreateAlias creates one stream alias.
func (c *ControllerV1) CreateAlias(ctx context.Context, req *v1.CreateAliasReq) (res *v1.CreateAliasRes, err error) {
	id, err := c.mediaSvc.CreateAlias(ctx, mediasvc.AliasMutationInput{
		Alias:      req.Alias,
		AutoRemove: req.AutoRemove,
		StreamPath: req.StreamPath,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateAliasRes{Id: id}, nil
}
