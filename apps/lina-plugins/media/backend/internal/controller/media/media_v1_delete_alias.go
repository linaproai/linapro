// This file implements the stream alias deletion controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
)

// DeleteAlias deletes one stream alias.
func (c *ControllerV1) DeleteAlias(ctx context.Context, req *v1.DeleteAliasReq) (res *v1.DeleteAliasRes, err error) {
	if err = c.mediaSvc.DeleteAlias(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteAliasRes{Id: req.Id}, nil
}
