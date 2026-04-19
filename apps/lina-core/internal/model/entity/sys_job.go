// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysJob is the golang structure for table sys_job.
type SysJob struct {
	Id                   uint64      `json:"id"                   orm:"id"                     description:"任务ID"`
	GroupId              uint64      `json:"groupId"              orm:"group_id"               description:"所属分组ID"`
	Name                 string      `json:"name"                 orm:"name"                   description:"任务名称"`
	Description          string      `json:"description"          orm:"description"            description:"任务描述"`
	TaskType             string      `json:"taskType"             orm:"task_type"              description:"任务类型（handler/shell）"`
	HandlerRef           string      `json:"handlerRef"           orm:"handler_ref"            description:"Handler 唯一引用"`
	Params               string      `json:"params"               orm:"params"                 description:"Handler 参数JSON"`
	TimeoutSeconds       int         `json:"timeoutSeconds"       orm:"timeout_seconds"        description:"执行超时时间（秒）"`
	ShellCmd             string      `json:"shellCmd"             orm:"shell_cmd"              description:"Shell 脚本内容"`
	WorkDir              string      `json:"workDir"              orm:"work_dir"               description:"工作目录"`
	Env                  string      `json:"env"                  orm:"env"                    description:"环境变量JSON"`
	CronExpr             string      `json:"cronExpr"             orm:"cron_expr"              description:"Cron 表达式"`
	Timezone             string      `json:"timezone"             orm:"timezone"               description:"时区标识"`
	Scope                string      `json:"scope"                orm:"scope"                  description:"调度范围（master_only/all_node）"`
	Concurrency          string      `json:"concurrency"          orm:"concurrency"            description:"并发策略（singleton/parallel）"`
	MaxConcurrency       int         `json:"maxConcurrency"       orm:"max_concurrency"        description:"并发上限"`
	MaxExecutions        int         `json:"maxExecutions"        orm:"max_executions"         description:"最大执行次数（0=无限）"`
	ExecutedCount        int64       `json:"executedCount"        orm:"executed_count"         description:"已执行次数"`
	StopReason           string      `json:"stopReason"           orm:"stop_reason"            description:"停止原因"`
	LogRetentionOverride string      `json:"logRetentionOverride" orm:"log_retention_override" description:"日志保留策略覆盖JSON"`
	Status               string      `json:"status"               orm:"status"                 description:"任务状态（enabled/disabled/paused_by_plugin）"`
	IsBuiltin            int         `json:"isBuiltin"            orm:"is_builtin"             description:"是否内置任务（1=是 0=否）"`
	SeedVersion          int         `json:"seedVersion"          orm:"seed_version"           description:"种子版本号"`
	CreatedBy            int64       `json:"createdBy"            orm:"created_by"             description:"创建者用户ID"`
	UpdatedBy            int64       `json:"updatedBy"            orm:"updated_by"             description:"更新者用户ID"`
	CreatedAt            *gtime.Time `json:"createdAt"            orm:"created_at"             description:"创建时间"`
	UpdatedAt            *gtime.Time `json:"updatedAt"            orm:"updated_at"             description:"更新时间"`
	DeletedAt            *gtime.Time `json:"deletedAt"            orm:"deleted_at"             description:"删除时间"`
}
