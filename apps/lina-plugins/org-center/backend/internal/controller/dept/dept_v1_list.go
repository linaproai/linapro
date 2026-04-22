// dept_v1_list.go implements the controller method that returns department list data.

package dept

import (
	"context"

	v1 "lina-plugin-org-center/backend/api/dept/v1"
	deptsvc "lina-plugin-org-center/backend/internal/service/dept"
)

// List returns department list.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.deptSvc.List(ctx, deptsvc.ListInput{Name: req.Name, Status: req.Status})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.DeptEntity, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, toAPIDeptEntity(item))
	}
	return &v1.ListRes{List: items}, nil
}
