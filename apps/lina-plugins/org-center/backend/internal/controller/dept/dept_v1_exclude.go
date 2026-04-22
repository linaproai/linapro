// dept_v1_exclude.go implements the controller method that excludes one subtree.

package dept

import (
	"context"

	v1 "lina-plugin-org-center/backend/api/dept/v1"
	deptsvc "lina-plugin-org-center/backend/internal/service/dept"
)

// Exclude returns department list excluding the specified node.
func (c *ControllerV1) Exclude(ctx context.Context, req *v1.ExcludeReq) (res *v1.ExcludeRes, err error) {
	list, err := c.deptSvc.Exclude(ctx, deptsvc.ExcludeInput{Id: req.Id})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.DeptEntity, 0, len(list))
	for _, item := range list {
		items = append(items, toAPIDeptEntity(item))
	}
	return &v1.ExcludeRes{List: items}, nil
}
