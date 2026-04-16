package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// OperLog Clean API

// CleanReq defines the request for clearing operation logs.
type CleanReq struct {
	g.Meta    `path:"/operlog/clean" method:"delete" tags:"操作日志" summary:"清空操作日志" dc:"清空指定时间范围内的操作日志，不传时间则清空全部" permission:"monitor:operlog:clear"`
	BeginTime string `json:"beginTime" dc:"清理起始时间" eg:"2025-01-01"`
	EndTime   string `json:"endTime" dc:"清理截止时间" eg:"2025-06-30"`
}

// CleanRes Operation log clean response
type CleanRes struct {
	Deleted int `json:"deleted" dc:"实际删除的记录数" eg:"1000"`
}
