package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateReq defines the request for updating one scheduled job.
type UpdateReq struct {
	g.Meta `path:"/job/{id}" method:"put" tags:"Job Scheduling / Scheduled Jobs" summary:"update task" operLog:"update" dc:"Update the user-built Shell scheduled job configuration based on the task ID; source code registration tasks are only allowed to be viewed and triggered, and are not allowed to be modified through the public interface." permission:"system:job:edit"`
	Id     int64 `json:"id" v:"required" dc:"Task ID" eg:"1"`
	JobPayload
}

// UpdateRes defines the response for updating one scheduled job.
type UpdateRes struct{}
