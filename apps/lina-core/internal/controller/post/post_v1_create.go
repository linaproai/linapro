package post

import (
	"context"

	v1 "lina-core/api/post/v1"
	postsvc "lina-core/internal/service/post"
)

// Create creates a post
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	sort := 0
	if req.Sort != nil {
		sort = *req.Sort
	}
	status := 1
	if req.Status != nil {
		status = *req.Status
	}
	id, err := c.postSvc.Create(ctx, postsvc.CreateInput{
		DeptId: req.DeptId,
		Code:   req.Code,
		Name:   req.Name,
		Sort:   sort,
		Status: status,
		Remark: req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Id: id}, nil
}
