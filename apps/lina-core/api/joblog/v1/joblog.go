// Package v1 defines shared scheduled-job log API DTOs and compact enum contracts.
package v1

// Trigger identifies how one job execution was started.
type Trigger string

// Supported execution trigger types.
const (
	TriggerCron   Trigger = "cron"
	TriggerManual Trigger = "manual"
)

// Status identifies the recorded execution outcome.
type Status string

// Supported execution log statuses.
const (
	StatusRunning               Status = "running"
	StatusSuccess               Status = "success"
	StatusFailed                Status = "failed"
	StatusCancelled             Status = "cancelled"
	StatusTimeout               Status = "timeout"
	StatusSkippedNotPrimary     Status = "skipped_not_primary"
	StatusSkippedSingleton      Status = "skipped_singleton"
	StatusSkippedMaxConcurrency Status = "skipped_max_concurrency"
)

// JobLogItem exposes scheduled-job execution log fields needed by the UI.
type JobLogItem struct {
	Id             int64   `json:"id" dc:"Log ID" eg:"1001"`
	JobId          int64   `json:"jobId" dc:"Owning job ID" eg:"1"`
	JobSnapshot    string  `json:"jobSnapshot" dc:"Job snapshot JSON at execution time" eg:"{}"`
	NodeId         string  `json:"nodeId" dc:"Execution node identifier" eg:"node-a"`
	Trigger        Trigger `json:"trigger" dc:"Trigger type" eg:"manual"`
	ParamsSnapshot string  `json:"paramsSnapshot" dc:"Parameter snapshot JSON at execution time" eg:"{}"`
	StartAt        *int64  `json:"startAt" dc:"Start time as Unix timestamp in milliseconds" eg:"1778733600000"`
	EndAt          *int64  `json:"endAt" dc:"End time as Unix timestamp in milliseconds" eg:"1778733660000"`
	DurationMs     int64   `json:"durationMs" dc:"Execution duration in milliseconds" eg:"1000"`
	Status         Status  `json:"status" dc:"Execution status" eg:"success"`
	ErrMsg         string  `json:"errMsg" dc:"Error summary" eg:""`
	ResultJson     string  `json:"resultJson" dc:"Execution result JSON" eg:"{}"`
	CreatedAt      *int64  `json:"createdAt" dc:"Creation time as Unix timestamp in milliseconds" eg:"1778733600000"`
}
