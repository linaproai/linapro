package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteReq defines the request for deleting scheduled jobs.
type DeleteReq struct {
	g.Meta `path:"/job/{ids}" method:"delete" tags:"任务调度/定时任务" summary:"删除任务" dc:"按任务ID批量删除定时任务，内置任务不允许删除" permission:"system:job:remove"`
	Ids    string `json:"ids" v:"required" dc:"任务ID，多个用逗号分隔" eg:"1,2,3"`
}

// DeleteRes defines the response for deleting scheduled jobs.
type DeleteRes struct{}
