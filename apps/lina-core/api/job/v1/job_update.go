package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateReq defines the request for updating one scheduled job.
type UpdateReq struct {
	g.Meta `path:"/job/{id}" method:"put" tags:"任务调度/定时任务" summary:"更新任务" operLog:"update" dc:"根据任务ID更新用户自建 Shell 定时任务配置；源码注册任务只允许查看和触发，不允许通过公共接口修改" permission:"system:job:edit"`
	Id     uint64 `json:"id" v:"required" dc:"任务ID" eg:"1"`
	JobPayload
}

// UpdateRes defines the response for updating one scheduled job.
type UpdateRes struct{}
