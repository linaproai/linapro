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
	Id             any         // 日志ID
	JobId          any         // 所属任务ID
	JobSnapshot    any         // 执行时任务快照JSON
	NodeId         any         // 执行节点标识
	Trigger        any         // 触发方式（cron/manual）
	ParamsSnapshot any         // 执行时参数快照JSON
	StartAt        *gtime.Time // 开始时间
	EndAt          *gtime.Time // 结束时间
	DurationMs     any         // 执行耗时（毫秒）
	Status         any         // 执行状态
	ErrMsg         any         // 错误摘要
	ResultJson     any         // 执行结果JSON
	CreatedAt      *gtime.Time // 创建时间
}
