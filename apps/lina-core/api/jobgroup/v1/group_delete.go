package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteReq defines the request for deleting scheduled job groups.
type DeleteReq struct {
	g.Meta `path:"/job-group/{ids}" method:"delete" tags:"任务调度/任务分组" summary:"删除分组" dc:"按分组ID批量删除任务分组，默认分组不允许删除，其它分组下任务需迁移到默认分组" permission:"system:jobgroup:remove"`
	Ids    string `json:"ids" v:"required" dc:"分组ID，多个用逗号分隔" eg:"2,3"`
}

// DeleteRes defines the response for deleting scheduled job groups.
type DeleteRes struct{}
