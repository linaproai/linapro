package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateStatusReq defines the request for updating one scheduled job status.
type UpdateStatusReq struct {
	g.Meta `path:"/job/{id}/status" method:"put" tags:"Job Scheduling / Scheduled Jobs" summary:"Update task status" dc:"Enable or disable scheduled jobs created by specified users; source code registration tasks are read-only definitions and are not allowed to be started or stopped through the public interface." permission:"system:job:status"`
	Id     uint64 `json:"id" v:"required" dc:"Task ID" eg:"1"`
	Status string `json:"status" v:"required|in:enabled,disabled" dc:"Task status: enabled=enabled disabled=disabled" eg:"enabled"`
}

// UpdateStatusRes defines the response for updating one scheduled job status.
type UpdateStatusRes struct{}
