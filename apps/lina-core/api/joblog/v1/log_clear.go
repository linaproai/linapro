package v1

import "github.com/gogf/gf/v2/frame/g"

// ClearReq defines the request for clearing scheduled job logs.
type ClearReq struct {
	g.Meta `path:"/job/log" method:"delete" tags:"任务调度/执行日志" summary:"清理执行日志" dc:"支持按日志ID批量删除、按任务ID清空日志，或在不传 jobId 和 logIds 时清空全部执行日志" permission:"system:joblog:remove"`
	JobId  *uint64 `json:"jobId" dc:"待清空日志的任务ID；与 logIds 二选一，不传则可清空全部任务日志" eg:"1"`
	LogIds string  `json:"logIds" dc:"待批量删除的日志ID列表，使用逗号分隔；传入后优先按指定日志删除" eg:"1,2,3"`
}

// ClearRes defines the response for clearing scheduled job logs.
type ClearRes struct{}
