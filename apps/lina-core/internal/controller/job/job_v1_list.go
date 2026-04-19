// This file implements the scheduled job list endpoint.

package job

import (
	"context"

	"lina-core/api/job/v1"
	"lina-core/internal/service/jobmeta"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
)

// List handles scheduled job list requests.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.jobMgmtSvc.ListJobs(ctx, jobmgmtsvc.ListJobsInput{
		PageNum:        req.PageNum,
		PageSize:       req.PageSize,
		GroupID:        req.GroupId,
		Status:         jobmeta.NormalizeJobStatus(req.Status),
		TaskType:       jobmeta.NormalizeTaskType(req.TaskType),
		Keyword:        req.Keyword,
		Scope:          jobmeta.NormalizeJobScope(req.Scope),
		Concurrency:    jobmeta.NormalizeJobConcurrency(req.Concurrency),
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
			SysJob:    item.SysJob,
			GroupCode: item.GroupCode,
			GroupName: item.GroupName,
		})
	}
	return &v1.ListRes{List: items, Total: out.Total}, nil
}
