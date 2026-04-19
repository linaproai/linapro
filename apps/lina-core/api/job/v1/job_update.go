package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateReq defines the request for updating one scheduled job.
type UpdateReq struct {
	g.Meta `path:"/job/{id}" method:"put" tags:"定时任务管理" summary:"更新任务" operLog:"update" dc:"根据任务ID更新定时任务配置，内置任务仅允许修改开放字段" permission:"system:job:edit"`
	Id     uint64 `json:"id" v:"required" dc:"任务ID" eg:"1"`
	JobPayload
}

// UpdateRes defines the response for updating one scheduled job.
type UpdateRes struct{}
