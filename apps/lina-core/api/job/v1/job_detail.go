package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DetailReq defines the request for querying one scheduled job detail.
type DetailReq struct {
	g.Meta `path:"/job/{id}" method:"get" tags:"Job Scheduling / Scheduled Jobs" summary:"Get task details" dc:"Query scheduled job details based on task ID and return basic configuration, grouping information and current policy fields" permission:"system:job:list"`
	Id     int64 `json:"id" v:"required" dc:"Task ID" eg:"1"`
}

// DetailRes defines the response for querying one scheduled job detail.
type DetailRes struct {
	*entity.SysJob `dc:"Job details" eg:""`
	GroupCode      string `json:"groupCode" dc:"The group code to which it belongs" eg:"default"`
	GroupName      string `json:"groupName" dc:"Group name" eg:"Default grouping"`
}
