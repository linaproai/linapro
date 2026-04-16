package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// LoginLog Clean API

// CleanReq defines the request for clearing login logs.
type CleanReq struct {
	g.Meta    `path:"/loginlog/clean" method:"delete" tags:"登录日志" summary:"清空登录日志" dc:"清空指定时间范围内的登录日志，不传时间则清空全部" permission:"monitor:loginlog:clear"`
	BeginTime string `json:"beginTime" dc:"清理起始时间" eg:"2025-01-01"`
	EndTime   string `json:"endTime" dc:"清理截止时间" eg:"2025-06-30"`
}

// CleanRes Login log clean response
type CleanRes struct {
	Deleted int `json:"deleted" dc:"实际删除的记录数" eg:"500"`
}
