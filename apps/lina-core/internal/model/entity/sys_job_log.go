// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysJobLog is the golang structure for table sys_job_log.
type SysJobLog struct {
	Id             int64       `json:"id"             orm:"id"              description:"Log ID"`
	TenantId       int         `json:"tenantId"       orm:"tenant_id"       description:"Owning tenant ID, 0 means PLATFORM"`
	JobId          int64       `json:"jobId"          orm:"job_id"          description:"Owning job ID"`
	JobSnapshot    string      `json:"jobSnapshot"    orm:"job_snapshot"    description:"Job snapshot JSON at execution time"`
	NodeId         string      `json:"nodeId"         orm:"node_id"         description:"Execution node identifier"`
	Trigger        string      `json:"trigger"        orm:"trigger"         description:"Trigger type: cron/manual"`
	ParamsSnapshot string      `json:"paramsSnapshot" orm:"params_snapshot" description:"Parameter snapshot JSON at execution time"`
	StartAt        *gtime.Time `json:"startAt"        orm:"start_at"        description:"Start time"`
	EndAt          *gtime.Time `json:"endAt"          orm:"end_at"          description:"End time"`
	DurationMs     int64       `json:"durationMs"     orm:"duration_ms"     description:"Execution duration in milliseconds"`
	Status         string      `json:"status"         orm:"status"          description:"Execution status"`
	ErrMsg         string      `json:"errMsg"         orm:"err_msg"         description:"Error summary"`
	ResultJson     string      `json:"resultJson"     orm:"result_json"     description:"Execution result JSON"`
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"      description:"Creation time"`
}
