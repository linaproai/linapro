package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateStatusReq defines the request for updating one scheduled job status.
type UpdateStatusReq struct {
	g.Meta `path:"/job/{id}/status" method:"put" tags:"任务调度/定时任务" summary:"更新任务状态" dc:"启用或停用指定用户自建定时任务；源码注册任务为只读定义，不允许通过公共接口启停" permission:"system:job:status"`
	Id     uint64 `json:"id" v:"required" dc:"任务ID" eg:"1"`
	Status string `json:"status" v:"required|in:enabled,disabled" dc:"任务状态：enabled=启用 disabled=停用" eg:"enabled"`
}

// UpdateStatusRes defines the response for updating one scheduled job status.
type UpdateStatusRes struct{}
