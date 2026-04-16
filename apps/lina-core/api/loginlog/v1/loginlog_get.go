package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// LoginLog Get API

// GetReq defines the request for retrieving login log details.
type GetReq struct {
	g.Meta `path:"/loginlog/{id}" method:"get" tags:"登录日志" summary:"获取登录日志详情" dc:"根据日志ID获取登录日志的详细信息，包括登录IP、浏览器、操作系统等" permission:"monitor:loginlog:query"`
	Id     int `json:"id" v:"required" dc:"登录日志ID" eg:"1"`
}

// GetRes Login log detail response
type GetRes struct {
	*entity.SysLoginLog
}
