package v1

import "github.com/gogf/gf/v2/frame/g"

// ResetReq defines the request for resetting one scheduled job execution counter.
type ResetReq struct {
	g.Meta `path:"/job/{id}/reset" method:"post" tags:"Job Scheduling / Scheduled Jobs" summary:"Reset execution count" dc:"Reset the executed_count of the specified user-created task to 0, without affecting the historical execution log; source code registration tasks are not allowed to be reset through the public interface" permission:"system:job:reset"`
	Id     int64 `json:"id" v:"required" dc:"Task ID" eg:"1"`
}

// ResetRes defines the response for resetting one scheduled job execution counter.
type ResetRes struct{}
