package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// OperLog Get API

// GetReq defines the request for retrieving operation log details.
type GetReq struct {
	g.Meta `path:"/operlog/{id}" method:"get" tags:"操作日志" summary:"获取操作日志详情" dc:"根据日志ID获取操作日志的详细信息，包括请求参数、响应结果、耗时等" permission:"monitor:operlog:query"`
	Id     int `json:"id" v:"required" dc:"操作日志ID" eg:"1"`
}

// GetRes Operation log detail response
type GetRes struct {
	*entity.SysOperLog
}
