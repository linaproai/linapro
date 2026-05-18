// This file implements the scheduled job group list endpoint.

package jobgroup

import (
	"context"

	"lina-core/api/jobgroup/v1"
	"lina-core/internal/model/entity"
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
			JobGroupItem: jobGroupItem(item.SysJobGroup),
			JobCount:     item.JobCount,
		})
	}
	return &v1.ListRes{List: items, Total: out.Total}, nil
}

// jobGroupItem maps a scheduled-job group entity to the API-safe response DTO.
func jobGroupItem(group *entity.SysJobGroup) v1.JobGroupItem {
	if group == nil {
		return v1.JobGroupItem{}
	}
	return v1.JobGroupItem{
		Id:        group.Id,
		Code:      group.Code,
		Name:      group.Name,
		Remark:    group.Remark,
		SortOrder: group.SortOrder,
		IsDefault: group.IsDefault,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}
}
