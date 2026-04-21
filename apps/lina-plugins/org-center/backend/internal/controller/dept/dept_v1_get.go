// dept_v1_get.go implements the controller method that returns department detail.

package dept

import (
	"context"

	v1 "lina-plugin-org-center/backend/api/dept/v1"
)

// Get returns department details by ID.
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	deptItem, err := c.deptSvc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetRes{DeptEntity: toAPIDeptEntity(deptItem)}, nil
}
