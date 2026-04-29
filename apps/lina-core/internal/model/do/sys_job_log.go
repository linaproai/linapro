// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysJobLog is the golang structure of table sys_job_log for DAO operations like Where/Data.
type SysJobLog struct {
	g.Meta         `orm:"table:sys_job_log, do:true"`
	Id             any         // Log ID
	JobId          any         // Owning job ID
	JobSnapshot    any         // Job snapshot JSON at execution time
	NodeId         any         // Execution node identifier
	Trigger        any         // Trigger type: cron/manual
	ParamsSnapshot any         // Parameter snapshot JSON at execution time
	StartAt        *gtime.Time // Start time
	EndAt          *gtime.Time // End time
	DurationMs     any         // Execution duration in milliseconds
	Status         any         // Execution status
	ErrMsg         any         // Error summary
	ResultJson     any         // Execution result JSON
	CreatedAt      *gtime.Time // Creation time
}
