package v1

import "github.com/gogf/gf/v2/frame/g"

// GroupPayload defines the shared mutable fields for scheduled job groups.
type GroupPayload struct {
	Code      string `json:"code" v:"required|length:1,64" dc:"分组编码，全局唯一" eg:"default"`
	Name      string `json:"name" v:"required|length:1,128" dc:"分组名称" eg:"默认分组"`
	Remark    string `json:"remark" dc:"分组备注" eg:"系统默认任务分组"`
	SortOrder int    `json:"sortOrder" d:"0" dc:"显示排序，数值越小越靠前" eg:"0"`
}

// CreateReq defines the request for creating one scheduled job group.
type CreateReq struct {
	g.Meta `path:"/job-group" method:"post" tags:"任务调度/任务分组" summary:"创建分组" dc:"创建一个新的定时任务分组，供任务管理页面按分组组织任务" permission:"system:jobgroup:add"`
	GroupPayload
}

// CreateRes defines the response for creating one scheduled job group.
type CreateRes struct {
	Id uint64 `json:"id" dc:"新建分组ID" eg:"1"`
}
