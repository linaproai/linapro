package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DetailReq defines the request for querying one scheduled job log detail.
type DetailReq struct {
	g.Meta `path:"/job/log/{id}" method:"get" tags:"任务调度/执行日志" summary:"获取日志详情" dc:"根据日志ID查询任务执行日志详情，返回任务快照、执行结果和错误摘要" permission:"system:joblog:list"`
	Id     uint64 `json:"id" v:"required" dc:"日志ID" eg:"1001"`
}

// DetailRes defines the response for querying one scheduled job log detail.
type DetailRes struct {
	*entity.SysJobLog `dc:"执行日志详情" eg:""`
	JobName           string `json:"jobName" dc:"任务名称" eg:"任务日志清理"`
}
