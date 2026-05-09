package v1

import "github.com/gogf/gf/v2/frame/g"

// ClearReq defines the request for clearing scheduled job logs.
type ClearReq struct {
	g.Meta `path:"/job/log" method:"delete" tags:"Job Scheduling / Execution Logs" summary:"Clean execution log" dc:"Supports batch deletion by log ID, clearing logs by task ID, or clearing all execution logs when jobId and logIds are not passed." permission:"system:joblog:remove"`
	JobId  *int64 `json:"jobId" dc:"The task ID of the log to be cleared; choose one of the two logIds, if not passed, all task logs can be cleared" eg:"1"`
	LogIds string `json:"logIds" dc:"List of log IDs to be deleted in batches, separated by commas; after passing in, priority will be given to deleting the specified logs." eg:"1,2,3"`
}

// ClearRes defines the response for clearing scheduled job logs.
type ClearRes struct{}
