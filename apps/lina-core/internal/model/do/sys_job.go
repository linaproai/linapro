// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysJob is the golang structure of table sys_job for DAO operations like Where/Data.
type SysJob struct {
	g.Meta               `orm:"table:sys_job, do:true"`
	Id                   any         // Job ID
	TenantId             any         // Owning tenant ID, 0 means PLATFORM
	GroupId              any         // Owning group ID
	Name                 any         // Job name
	Description          any         // Job description
	TaskType             any         // Job type: handler/shell
	HandlerRef           any         // Unique handler reference
	Params               any         // Handler parameters JSON
	TimeoutSeconds       any         // Execution timeout in seconds
	ShellCmd             any         // Shell script content
	WorkDir              any         // Working directory
	Env                  any         // Environment variables JSON
	CronExpr             any         // Cron expression
	Timezone             any         // Timezone identifier
	Scope                any         // Scheduling scope: master_only/all_node
	Concurrency          any         // Concurrency policy: singleton/parallel
	MaxConcurrency       any         // Maximum concurrency
	MaxExecutions        any         // Maximum executions, 0 means unlimited
	ExecutedCount        any         // Executed count
	StopReason           any         // Stop reason
	LogRetentionOverride any         // Log retention override JSON
	Status               any         // Job status: enabled/disabled/paused_by_plugin
	IsBuiltin            any         // Built-in job flag: 1=yes, 0=no
	SeedVersion          any         // Seed version number
	CreatedBy            any         // Creator user ID
	UpdatedBy            any         // Updater user ID
	CreatedAt            *gtime.Time // Creation time
	UpdatedAt            *gtime.Time // Update time
	DeletedAt            *gtime.Time // Deletion time
}
