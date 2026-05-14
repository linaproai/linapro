// This file defines shared scheduled-job response DTOs for the job API.
package v1

import "github.com/gogf/gf/v2/os/gtime"

// JobItem exposes scheduled-job fields needed by the management UI.
type JobItem struct {
	Id                   int64       `json:"id" dc:"Job ID" eg:"1"`
	GroupId              int64       `json:"groupId" dc:"Owning group ID" eg:"1"`
	Name                 string      `json:"name" dc:"Job name" eg:"Log cleanup"`
	Description          string      `json:"description" dc:"Job description" eg:"Clean expired logs"`
	TaskType             string      `json:"taskType" dc:"Job type: handler or shell" eg:"handler"`
	HandlerRef           string      `json:"handlerRef" dc:"Handler reference" eg:"host:cleanup-logs"`
	Params               string      `json:"params" dc:"Handler parameters JSON" eg:"{}"`
	TimeoutSeconds       int         `json:"timeoutSeconds" dc:"Execution timeout in seconds" eg:"30"`
	ShellCmd             string      `json:"shellCmd" dc:"Shell script content" eg:"echo hello"`
	WorkDir              string      `json:"workDir" dc:"Shell working directory" eg:"/tmp"`
	Env                  string      `json:"env" dc:"Shell environment JSON" eg:"{}"`
	CronExpr             string      `json:"cronExpr" dc:"Cron expression" eg:"0 0 * * * *"`
	Timezone             string      `json:"timezone" dc:"Timezone identifier" eg:"Asia/Shanghai"`
	Scope                string      `json:"scope" dc:"Scheduling scope" eg:"master_only"`
	Concurrency          string      `json:"concurrency" dc:"Concurrency policy" eg:"singleton"`
	MaxConcurrency       int         `json:"maxConcurrency" dc:"Maximum concurrency" eg:"1"`
	MaxExecutions        int         `json:"maxExecutions" dc:"Maximum executions, 0 means unlimited" eg:"0"`
	ExecutedCount        int64       `json:"executedCount" dc:"Executed count" eg:"12"`
	StopReason           string      `json:"stopReason" dc:"Stop reason" eg:""`
	LogRetentionOverride string      `json:"logRetentionOverride" dc:"Log retention override JSON" eg:"{\"mode\":\"days\",\"value\":30}"`
	Status               string      `json:"status" dc:"Job status" eg:"enabled"`
	IsBuiltin            int         `json:"isBuiltin" dc:"Built-in job flag: 1=yes 0=no" eg:"0"`
	SeedVersion          int         `json:"seedVersion" dc:"Built-in seed version" eg:"1"`
	CreatedBy            int64       `json:"createdBy" dc:"Creator user ID" eg:"1"`
	UpdatedBy            int64       `json:"updatedBy" dc:"Updater user ID" eg:"1"`
	CreatedAt            *gtime.Time `json:"createdAt" dc:"Creation time" eg:"2026-05-14 10:00:00"`
	UpdatedAt            *gtime.Time `json:"updatedAt" dc:"Update time" eg:"2026-05-14 10:00:00"`
}
