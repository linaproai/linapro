package v1

import "github.com/gogf/gf/v2/frame/g"

// CancelReq defines the request for cancelling one running scheduled job log.
type CancelReq struct {
	g.Meta `path:"/job/log/{id}/cancel" method:"post" tags:"任务调度/执行日志" summary:"终止运行实例" operLog:"other" dc:"终止指定运行中的任务实例；Shell 实例还需通过 system:job:shell 的附加权限校验" permission:"system:joblog:cancel"`
	Id     uint64 `json:"id" v:"required" dc:"日志ID" eg:"1001"`
}

// CancelRes defines the response for cancelling one running scheduled job log.
type CancelRes struct{}
