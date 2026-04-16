package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// LoginLog Export API

// ExportReq defines the request for exporting login logs.
type ExportReq struct {
	g.Meta         `path:"/loginlog/export" method:"get" tags:"登录日志" summary:"导出登录日志" operLog:"export" dc:"导出登录日志数据为Excel文件，支持按条件筛选导出，也支持导出指定ID的记录" permission:"monitor:loginlog:export"`
	UserName       string `json:"userName" dc:"按用户名筛选（模糊匹配）" eg:"admin"`
	Ip             string `json:"ip" dc:"按IP地址筛选（模糊匹配）" eg:"192.168"`
	Status         *int   `json:"status" dc:"按状态筛选：1=成功 0=失败" eg:"1"`
	BeginTime      string `json:"beginTime" dc:"按登录时间起始筛选" eg:"2025-01-01"`
	EndTime        string `json:"endTime" dc:"按登录时间结束筛选" eg:"2025-12-31"`
	OrderBy        string `json:"orderBy" dc:"排序字段" eg:"login_time"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"排序方向：asc或desc" eg:"desc"`
	Ids            []int  `json:"ids" dc:"指定导出的记录ID列表，不传则导出全部符合条件的记录" eg:"[1,2,3]"`
}

// ExportRes Login log export response
type ExportRes struct{}
