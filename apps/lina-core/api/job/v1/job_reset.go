package v1

import "github.com/gogf/gf/v2/frame/g"

// ResetReq defines the request for resetting one scheduled job execution counter.
type ResetReq struct {
	g.Meta `path:"/job/{id}/reset" method:"post" tags:"定时任务管理" summary:"重置执行计数" dc:"将指定用户自建任务的 executed_count 重置为0，不影响历史执行日志；源码注册任务不允许通过公共接口重置" permission:"system:job:reset"`
	Id     uint64 `json:"id" v:"required" dc:"任务ID" eg:"1"`
}

// ResetRes defines the response for resetting one scheduled job execution counter.
type ResetRes struct{}
