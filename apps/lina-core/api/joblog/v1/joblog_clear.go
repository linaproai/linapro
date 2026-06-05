package v1

import "github.com/gogf/gf/v2/frame/g"

// ClearReq defines the request for clearing scheduled job logs.
type ClearReq struct {
	g.Meta    `path:"/job/log" method:"delete" tags:"Job Scheduling / Execution Logs" summary:"Clean execution log" dc:"Supports batch deletion by log ID, clearing logs by task ID, deleting logs within a specified start-time range, or clearing all visible execution logs when no filter is passed." permission:"system:joblog:remove"`
	JobId     *int64 `json:"jobId" dc:"The task ID of the logs to be cleared; logIds takes priority when both are passed; if no filter is passed, all visible task logs can be cleared" eg:"1"`
	LogIds    string `json:"logIds" dc:"List of log IDs to be deleted in batches, separated by commas; after passing in, priority will be given to deleting the specified logs." eg:"1,2,3"`
	BeginTime string `json:"beginTime" dc:"Cleanup start time by execution start_at lower bound; ignored when logIds is passed" eg:"2026-04-19"`
	EndTime   string `json:"endTime" dc:"Cleanup end time by execution start_at upper bound; date-only values include the whole day; ignored when logIds is passed" eg:"2026-04-19"`
}

// ClearRes defines the response for clearing scheduled job logs.
type ClearRes struct {
	Deleted int64 `json:"deleted" dc:"Number of execution logs actually deleted" eg:"500"`
}
