// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysJobLog is the golang structure for table sys_job_log.
type SysJobLog struct {
	Id             uint64      `json:"id"             orm:"id"              description:"日志ID"`
	JobId          uint64      `json:"jobId"          orm:"job_id"          description:"所属任务ID"`
	JobSnapshot    string      `json:"jobSnapshot"    orm:"job_snapshot"    description:"执行时任务快照JSON"`
	NodeId         string      `json:"nodeId"         orm:"node_id"         description:"执行节点标识"`
	Trigger        string      `json:"trigger"        orm:"trigger"         description:"触发方式（cron/manual）"`
	ParamsSnapshot string      `json:"paramsSnapshot" orm:"params_snapshot" description:"执行时参数快照JSON"`
	StartAt        *gtime.Time `json:"startAt"        orm:"start_at"        description:"开始时间"`
	EndAt          *gtime.Time `json:"endAt"          orm:"end_at"          description:"结束时间"`
	DurationMs     int64       `json:"durationMs"     orm:"duration_ms"     description:"执行耗时（毫秒）"`
	Status         string      `json:"status"         orm:"status"          description:"执行状态"`
	ErrMsg         string      `json:"errMsg"         orm:"err_msg"         description:"错误摘要"`
	ResultJson     string      `json:"resultJson"     orm:"result_json"     description:"执行结果JSON"`
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"      description:"创建时间"`
}
