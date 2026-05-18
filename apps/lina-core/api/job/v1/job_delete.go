package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteReq defines the request for deleting scheduled jobs.
type DeleteReq struct {
	g.Meta `path:"/job/{ids}" method:"delete" tags:"Job Scheduling / Scheduled Jobs" summary:"Delete task" dc:"Delete scheduled jobs in batches by task ID. Built-in tasks are not allowed to be deleted." permission:"system:job:remove"`
	Ids    string `json:"ids" v:"required" dc:"Task ID, multiple separated by commas" eg:"1,2,3"`
}

// DeleteRes defines the response for deleting scheduled jobs.
type DeleteRes struct{}
