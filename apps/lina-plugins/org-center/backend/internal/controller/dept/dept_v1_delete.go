// dept_v1_delete.go implements the controller method that deletes departments.

package dept

import (
	"context"

	v1 "lina-plugin-org-center/backend/api/dept/v1"
)

// Delete deletes a department by ID.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	return nil, c.deptSvc.Delete(ctx, req.Id)
}
