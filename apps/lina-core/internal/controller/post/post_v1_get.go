package post

import (
	"context"

	v1 "lina-core/api/post/v1"
)

// Get returns post details
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	post, err := c.postSvc.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetRes{SysPost: post}, nil
}
