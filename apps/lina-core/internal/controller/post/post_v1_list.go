package post

import (
	"context"

	v1 "lina-core/api/post/v1"
	postsvc "lina-core/internal/service/post"
)

// List queries post list
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.postSvc.List(ctx, postsvc.ListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		DeptId:   req.DeptId,
		Code:     req.Code,
		Name:     req.Name,
		Status:   req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ListRes{
		List:  out.List,
		Total: out.Total,
	}, nil
}
