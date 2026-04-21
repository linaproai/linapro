// post_v1_get.go implements the controller method that returns post detail.

package post

import (
	"context"

	v1 "lina-plugin-org-center/backend/api/post/v1"
)

// Get returns post details.
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	postItem, err := c.postSvc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetRes{PostEntity: toAPIPostEntity(postItem)}, nil
}
