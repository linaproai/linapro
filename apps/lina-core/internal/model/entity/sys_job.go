// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysJob is the golang structure for table sys_job.
type SysJob struct {
	Id                   int64       `json:"id"                   orm:"id"                     description:"Job ID"`
	TenantId             int         `json:"tenantId"             orm:"tenant_id"              description:"Owning tenant ID, 0 means PLATFORM"`
	GroupId              int64       `json:"groupId"              orm:"group_id"               description:"Owning group ID"`
	Name                 string      `json:"name"                 orm:"name"                   description:"Job name"`
	Description          string      `json:"description"          orm:"description"            description:"Job description"`
	TaskType             string      `json:"taskType"             orm:"task_type"              description:"Job type: handler/shell"`
	HandlerRef           string      `json:"handlerRef"           orm:"handler_ref"            description:"Unique handler reference"`
	Params               string      `json:"params"               orm:"params"                 description:"Handler parameters JSON"`
	TimeoutSeconds       int         `json:"timeoutSeconds"       orm:"timeout_seconds"        description:"Execution timeout in seconds"`
	ShellCmd             string      `json:"shellCmd"             orm:"shell_cmd"              description:"Shell script content"`
	WorkDir              string      `json:"workDir"              orm:"work_dir"               description:"Working directory"`
	Env                  string      `json:"env"                  orm:"env"                    description:"Environment variables JSON"`
	CronExpr             string      `json:"cronExpr"             orm:"cron_expr"              description:"Cron expression"`
	Timezone             string      `json:"timezone"             orm:"timezone"               description:"Timezone identifier"`
	Scope                string      `json:"scope"                orm:"scope"                  description:"Scheduling scope: master_only/all_node"`
	Concurrency          string      `json:"concurrency"          orm:"concurrency"            description:"Concurrency policy: singleton/parallel"`
	MaxConcurrency       int         `json:"maxConcurrency"       orm:"max_concurrency"        description:"Maximum concurrency"`
	MaxExecutions        int         `json:"maxExecutions"        orm:"max_executions"         description:"Maximum executions, 0 means unlimited"`
	ExecutedCount        int64       `json:"executedCount"        orm:"executed_count"         description:"Executed count"`
	StopReason           string      `json:"stopReason"           orm:"stop_reason"            description:"Stop reason"`
	LogRetentionOverride string      `json:"logRetentionOverride" orm:"log_retention_override" description:"Log retention override JSON"`
	Status               string      `json:"status"               orm:"status"                 description:"Job status: enabled/disabled/paused_by_plugin"`
	IsBuiltin            int         `json:"isBuiltin"            orm:"is_builtin"             description:"Built-in job flag: 1=yes, 0=no"`
	SeedVersion          int         `json:"seedVersion"          orm:"seed_version"           description:"Seed version number"`
	CreatedBy            int64       `json:"createdBy"            orm:"created_by"             description:"Creator user ID"`
	UpdatedBy            int64       `json:"updatedBy"            orm:"updated_by"             description:"Updater user ID"`
	CreatedAt            *gtime.Time `json:"createdAt"            orm:"created_at"             description:"Creation time"`
	UpdatedAt            *gtime.Time `json:"updatedAt"            orm:"updated_at"             description:"Update time"`
	DeletedAt            *gtime.Time `json:"deletedAt"            orm:"deleted_at"             description:"Deletion time"`
}
