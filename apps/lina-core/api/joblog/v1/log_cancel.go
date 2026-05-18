package v1

import "github.com/gogf/gf/v2/frame/g"

// CancelReq defines the request for cancelling one running scheduled job log.
type CancelReq struct {
	g.Meta `path:"/job/log/{id}/cancel" method:"post" tags:"Job Scheduling / Execution Logs" summary:"Terminate running instance" operLog:"other" dc:"Terminate the specified running task instance; the Shell instance also needs to pass the additional permission verification of system:job:shell" permission:"system:joblog:cancel"`
	Id     int64 `json:"id" v:"required" dc:"Log ID" eg:"1001"`
}

// CancelRes defines the response for cancelling one running scheduled job log.
type CancelRes struct{}
