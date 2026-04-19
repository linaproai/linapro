// This file implements the scheduled job group list endpoint.

package jobgroup

import (
	"context"

	"lina-core/api/jobgroup/v1"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
)

// List handles scheduled job group list requests.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.jobMgmtSvc.ListGroups(ctx, jobmgmtsvc.ListGroupsInput{
		PageNum:        req.PageNum,
		PageSize:       req.PageSize,
		Code:           req.Code,
		Name:           req.Name,
		OrderBy:        req.OrderBy,
		OrderDirection: req.OrderDirection,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.ListItem, 0, len(out.List))
	for _, item := range out.List {
		if item == nil {
			continue
		}
		items = append(items, &v1.ListItem{
			SysJobGroup: item.SysJobGroup,
			JobCount:    item.JobCount,
		})
	}
	return &v1.ListRes{List: items, Total: out.Total}, nil
}
