// post_v1_delete.go implements the controller method that deletes posts.

package post

import (
	"context"

	v1 "lina-plugin-org-center/backend/api/post/v1"
)

// Delete deletes one or more posts.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	if err = c.postSvc.Delete(ctx, req.Ids); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}
