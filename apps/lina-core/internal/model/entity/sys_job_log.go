// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysJobLog is the golang structure for table sys_job_log.
type SysJobLog struct {
	Id         uint64      `json:"id"         orm:"id"          description:"日志ID"`
	JobId      uint64      `json:"jobId"      orm:"job_id"      description:"任务ID"`
	JobName    string      `json:"jobName"    orm:"job_name"    description:"任务名称"`
	Command    string      `json:"command"    orm:"command"     description:"执行指令"`
	Status     int         `json:"status"     orm:"status"      description:"执行状态：1=成功 0=失败"`
	StartTime  *gtime.Time `json:"startTime"  orm:"start_time"  description:"开始时间"`
	EndTime    *gtime.Time `json:"endTime"    orm:"end_time"    description:"结束时间"`
	Duration   int         `json:"duration"   orm:"duration"    description:"执行耗时(毫秒)"`
	ErrorMsg   string      `json:"errorMsg"   orm:"error_msg"   description:"错误信息"`
	CreateTime *gtime.Time `json:"createTime" orm:"create_time" description:"创建时间"`
}
