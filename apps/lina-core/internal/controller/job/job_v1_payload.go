// This file converts scheduled-job API payloads into service-layer inputs.

package job

import (
	"time"

	"lina-core/api/job/v1"
	"lina-core/internal/service/jobmeta"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
)

// buildSaveJobInput converts one API payload into a service-layer input.
func buildSaveJobInput(payload v1.JobPayload) jobmgmtsvc.SaveJobInput {
	var retentionOverride *jobmeta.RetentionOption
	if payload.LogRetentionOverride != nil {
		retentionOverride = &jobmeta.RetentionOption{
			Mode:  jobmeta.NormalizeRetentionMode(string(payload.LogRetentionOverride.Mode)),
			Value: payload.LogRetentionOverride.Value,
		}
	}
	return jobmgmtsvc.SaveJobInput{
		GroupID:              payload.GroupId,
		Name:                 payload.Name,
		Description:          payload.Description,
		TaskType:             jobmeta.NormalizeTaskType(string(payload.TaskType)),
		HandlerRef:           payload.HandlerRef,
		Params:               payload.Params,
		Timeout:              time.Duration(payload.TimeoutSeconds) * time.Second,
		ShellCmd:             payload.ShellCmd,
		WorkDir:              payload.WorkDir,
		Env:                  payload.Env,
		CronExpr:             payload.CronExpr,
		Timezone:             payload.Timezone,
		Scope:                jobmeta.NormalizeJobScope(string(payload.Scope)),
		Concurrency:          jobmeta.NormalizeJobConcurrency(string(payload.Concurrency)),
		MaxConcurrency:       payload.MaxConcurrency,
		MaxExecutions:        payload.MaxExecutions,
		Status:               jobmeta.NormalizeJobStatus(string(payload.Status)),
		LogRetentionOverride: retentionOverride,
	}
}
