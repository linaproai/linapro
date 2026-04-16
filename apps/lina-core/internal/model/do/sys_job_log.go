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
	g.Meta     `orm:"table:sys_job_log, do:true"`
	Id         any         // 日志ID
	JobId      any         // 任务ID
	JobName    any         // 任务名称
	Command    any         // 执行指令
	Status     any         // 执行状态：1=成功 0=失败
	StartTime  *gtime.Time // 开始时间
	EndTime    *gtime.Time // 结束时间
	Duration   any         // 执行耗时(毫秒)
	ErrorMsg   any         // 错误信息
	CreateTime *gtime.Time // 创建时间
}
