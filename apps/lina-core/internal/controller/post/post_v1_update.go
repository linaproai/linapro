package post

import (
	"context"

	v1 "lina-core/api/post/v1"
	postsvc "lina-core/internal/service/post"
)

// Update updates a post
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	err = c.postSvc.Update(ctx, postsvc.UpdateInput{
		Id:     req.Id,
		DeptId: req.DeptId,
		Code:   req.Code,
		Name:   req.Name,
		Sort:   req.Sort,
		Status: req.Status,
		Remark: req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}
