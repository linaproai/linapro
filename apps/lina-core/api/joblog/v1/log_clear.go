package v1

import "github.com/gogf/gf/v2/frame/g"

// ClearReq defines the request for clearing scheduled job logs.
type ClearReq struct {
	g.Meta `path:"/job/log" method:"delete" tags:"任务执行日志" summary:"清空执行日志" dc:"按任务ID清空执行日志；若不传 jobId 则清空全部日志" permission:"system:joblog:remove"`
	JobId  *uint64 `json:"jobId" dc:"待清空日志的任务ID；不传则清空全部任务日志" eg:"1"`
}

// ClearRes defines the response for clearing scheduled job logs.
type ClearRes struct{}
