// This file implements the scheduled job list endpoint.

package job

import (
	"context"

	v1 "lina-core/api/job/v1"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobmeta"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	"lina-core/pkg/apitime"
	"lina-core/pkg/statusflag"
)

// List handles scheduled job list requests.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.jobMgmtSvc.ListJobs(ctx, jobmgmtsvc.ListJobsInput{
		PageNum:        req.PageNum,
		PageSize:       req.PageSize,
		GroupID:        req.GroupId,
		Status:         jobmeta.NormalizeJobStatus(string(req.Status)),
		TaskType:       jobmeta.NormalizeTaskType(string(req.TaskType)),
		Keyword:        req.Keyword,
		Scope:          jobmeta.NormalizeJobScope(string(req.Scope)),
		Concurrency:    jobmeta.NormalizeJobConcurrency(string(req.Concurrency)),
		OrderBy:        req.OrderBy,
		OrderDirection: string(req.OrderDirection),
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
			JobItem:   jobItem(item.SysJob),
			GroupCode: item.GroupCode,
			GroupName: item.GroupName,
		})
	}
	return &v1.ListRes{List: items, Total: out.Total}, nil
}

// jobItem maps a scheduled-job entity to the API-safe response DTO.
func jobItem(job *entity.SysJob) v1.JobItem {
	if job == nil {
		return v1.JobItem{}
	}
	return v1.JobItem{
		Id:                   job.Id,
		GroupId:              job.GroupId,
		Name:                 job.Name,
		Description:          job.Description,
		TaskType:             v1.TaskType(job.TaskType),
		HandlerRef:           job.HandlerRef,
		Params:               job.Params,
		TimeoutSeconds:       job.TimeoutSeconds,
		ShellCmd:             job.ShellCmd,
		WorkDir:              job.WorkDir,
		Env:                  job.Env,
		CronExpr:             job.CronExpr,
		Timezone:             job.Timezone,
		Scope:                v1.Scope(job.Scope),
		Concurrency:          v1.Concurrency(job.Concurrency),
		MaxConcurrency:       job.MaxConcurrency,
		MaxExecutions:        job.MaxExecutions,
		ExecutedCount:        job.ExecutedCount,
		StopReason:           job.StopReason,
		LogRetentionOverride: job.LogRetentionOverride,
		Status:               v1.Status(job.Status),
		IsBuiltin:            statusflag.YesNo(job.IsBuiltin),
		SeedVersion:          job.SeedVersion,
		CreatedBy:            job.CreatedBy,
		UpdatedBy:            job.UpdatedBy,
		CreatedAt:            apitime.Milli(job.CreatedAt),
		UpdatedAt:            apitime.Milli(job.UpdatedAt),
	}
}
