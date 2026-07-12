package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteReq defines the request for deleting scheduled jobs.
type DeleteReq struct {
	g.Meta `path:"/job" method:"delete" tags:"Job Scheduling / Task Management" summary:"Delete task" dc:"Delete scheduled jobs in batches by task ID. Built-in tasks are not allowed to be deleted. Pass ids as a query array (ids[]=1&ids[]=2)." permission:"system:job:remove"`
	Ids    []int64 `json:"ids" v:"required|min-length:1" dc:"Task ID list as a query array, e.g. ids[]=1&ids[]=2&ids[]=3" eg:"[1,2,3]"`
}

// DeleteRes defines the response for deleting scheduled jobs.
type DeleteRes struct{}
