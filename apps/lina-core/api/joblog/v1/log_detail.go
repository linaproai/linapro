package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DetailReq defines the request for querying one scheduled job log detail.
type DetailReq struct {
	g.Meta `path:"/job/log/{id}" method:"get" tags:"Job Scheduling / Execution Logs" summary:"Get log details" dc:"Query the task execution log details based on the log ID and return the task snapshot, execution results and error summary" permission:"system:joblog:list"`
	Id     uint64 `json:"id" v:"required" dc:"Log ID" eg:"1001"`
}

// DetailRes defines the response for querying one scheduled job log detail.
type DetailRes struct {
	*entity.SysJobLog `dc:"Execution log details" eg:""`
	JobName           string `json:"jobName" dc:"Task name" eg:"Task log cleaning"`
}
