// post_v1_list.go implements the controller method that returns paged post data.

package post

import (
	"context"

	v1 "lina-plugin-org-center/backend/api/post/v1"
	postsvc "lina-plugin-org-center/backend/internal/service/post"
)

// List queries post list.
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
	items := make([]*v1.PostEntity, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, toAPIPostEntity(item))
	}
	return &v1.ListRes{List: items, Total: out.Total}, nil
}
