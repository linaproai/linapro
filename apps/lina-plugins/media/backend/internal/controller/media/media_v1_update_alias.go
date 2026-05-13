// This file implements the stream alias update controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// UpdateAlias updates one stream alias.
func (c *ControllerV1) UpdateAlias(ctx context.Context, req *v1.UpdateAliasReq) (res *v1.UpdateAliasRes, err error) {
	if err = c.mediaSvc.UpdateAlias(ctx, req.Id, mediasvc.AliasMutationInput{
		Alias:      req.Alias,
		AutoRemove: req.AutoRemove,
		StreamPath: req.StreamPath,
	}); err != nil {
		return nil, err
	}
	return &v1.UpdateAliasRes{Id: req.Id}, nil
}
