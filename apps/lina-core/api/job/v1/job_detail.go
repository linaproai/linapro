package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DetailReq defines the request for querying one scheduled job detail.
type DetailReq struct {
	g.Meta `path:"/job/{id}" method:"get" tags:"任务调度/定时任务" summary:"获取任务详情" dc:"根据任务ID查询定时任务详情，返回基础配置、分组信息以及当前策略字段" permission:"system:job:list"`
	Id     uint64 `json:"id" v:"required" dc:"任务ID" eg:"1"`
}

// DetailRes defines the response for querying one scheduled job detail.
type DetailRes struct {
	*entity.SysJob `dc:"任务详情" eg:""`
	GroupCode      string `json:"groupCode" dc:"所属分组编码" eg:"default"`
	GroupName      string `json:"groupName" dc:"所属分组名称" eg:"默认分组"`
}
