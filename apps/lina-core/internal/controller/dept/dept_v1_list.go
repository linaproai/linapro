package dept

import (
	"context"

	v1 "lina-core/api/dept/v1"
	deptsvc "lina-core/internal/service/dept"
)

// List returns department list.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.deptSvc.List(ctx, deptsvc.ListInput{
		Name:   req.Name,
		Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ListRes{
		List: out.List,
	}, nil
}
