package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateReq defines the request for updating one scheduled job group.
type UpdateReq struct {
	g.Meta `path:"/job-group/{id}" method:"put" tags:"任务分组管理" summary:"更新分组" dc:"根据分组ID更新定时任务分组信息，默认分组仅允许修改开放字段" permission:"system:jobgroup:edit"`
	Id     uint64 `json:"id" v:"required" dc:"分组ID" eg:"1"`
	GroupPayload
}

// UpdateRes defines the response for updating one scheduled job group.
type UpdateRes struct{}
