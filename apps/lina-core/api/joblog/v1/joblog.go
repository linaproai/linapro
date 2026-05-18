// This file defines shared scheduled-job log response DTOs for the joblog API.
package v1

import "github.com/gogf/gf/v2/os/gtime"

// JobLogItem exposes scheduled-job execution log fields needed by the UI.
type JobLogItem struct {
	Id             int64       `json:"id" dc:"Log ID" eg:"1001"`
	JobId          int64       `json:"jobId" dc:"Owning job ID" eg:"1"`
	JobSnapshot    string      `json:"jobSnapshot" dc:"Job snapshot JSON at execution time" eg:"{}"`
	NodeId         string      `json:"nodeId" dc:"Execution node identifier" eg:"node-a"`
	Trigger        string      `json:"trigger" dc:"Trigger type" eg:"manual"`
	ParamsSnapshot string      `json:"paramsSnapshot" dc:"Parameter snapshot JSON at execution time" eg:"{}"`
	StartAt        *gtime.Time `json:"startAt" dc:"Start time" eg:"2026-05-14 10:00:00"`
	EndAt          *gtime.Time `json:"endAt" dc:"End time" eg:"2026-05-14 10:01:00"`
	DurationMs     int64       `json:"durationMs" dc:"Execution duration in milliseconds" eg:"1000"`
	Status         string      `json:"status" dc:"Execution status" eg:"success"`
	ErrMsg         string      `json:"errMsg" dc:"Error summary" eg:""`
	ResultJson     string      `json:"resultJson" dc:"Execution result JSON" eg:"{}"`
	CreatedAt      *gtime.Time `json:"createdAt" dc:"Creation time" eg:"2026-05-14 10:00:00"`
}
