package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// LoginLog List API

// ListReq defines the request for listing login logs.
type ListReq struct {
	g.Meta         `path:"/loginlog" method:"get" tags:"登录日志" summary:"获取登录日志列表" dc:"分页查询登录日志列表，记录用户登录成功和失败的信息，支持多条件筛选和排序" permission:"monitor:loginlog:query"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	UserName       string `json:"userName" dc:"按用户名筛选（模糊匹配）" eg:"admin"`
	Ip             string `json:"ip" dc:"按IP地址筛选（模糊匹配）" eg:"192.168"`
	Status         *int   `json:"status" dc:"按状态筛选：1=成功 0=失败" eg:"1"`
	BeginTime      string `json:"beginTime" dc:"按登录时间起始筛选" eg:"2025-01-01"`
	EndTime        string `json:"endTime" dc:"按登录时间结束筛选" eg:"2025-12-31"`
	OrderBy        string `json:"orderBy" dc:"排序字段：id,login_time" eg:"login_time"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"排序方向：asc或desc" eg:"desc"`
}

// ListRes Login log list response
type ListRes struct {
	Items []*entity.SysLoginLog `json:"items" dc:"登录日志列表" eg:"[]"`
	Total int                   `json:"total" dc:"总条数" eg:"100"`
}
