package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateStatusReq defines the request for updating one scheduled job status.
type UpdateStatusReq struct {
	g.Meta `path:"/job/{id}/status" method:"put" tags:"定时任务管理" summary:"更新任务状态" dc:"启用或停用指定定时任务；paused_by_plugin 状态仅由系统内部维护" permission:"system:job:status"`
	Id     uint64 `json:"id" v:"required" dc:"任务ID" eg:"1"`
	Status string `json:"status" v:"required|in:enabled,disabled" dc:"任务状态：enabled=启用 disabled=停用" eg:"enabled"`
}

// UpdateStatusRes defines the response for updating one scheduled job status.
type UpdateStatusRes struct{}
