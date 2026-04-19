package v1

import "github.com/gogf/gf/v2/frame/g"

// TriggerReq defines the request for triggering one scheduled job manually.
type TriggerReq struct {
	g.Meta `path:"/job/{id}/trigger" method:"post" tags:"定时任务管理" summary:"手动触发任务" operLog:"other" dc:"立即触发指定任务执行一次，记录 trigger=manual 的执行日志" permission:"system:job:trigger"`
	Id     uint64 `json:"id" v:"required" dc:"任务ID" eg:"1"`
}

// TriggerRes defines the response for triggering one scheduled job manually.
type TriggerRes struct {
	LogId uint64 `json:"logId" dc:"本次执行生成的日志ID" eg:"1001"`
}
