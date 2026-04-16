package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Config List API

// ListReq defines the request for querying config list.
type ListReq struct {
	g.Meta    `path:"/config" method:"get" tags:"参数设置" summary:"获取参数设置列表" dc:"分页查询参数设置列表，支持按参数名称、参数键名模糊匹配，以及按创建时间范围筛选" permission:"system:config:query"`
	PageNum   int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize  int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Name      string `json:"name" dc:"按参数名称筛选（模糊匹配），不传则查询全部" eg:"主框架页"`
	Key       string `json:"key" dc:"按参数键名筛选（模糊匹配），不传则查询全部" eg:"sys.index"`
	BeginTime string `json:"beginTime" dc:"创建时间范围-开始时间，格式YYYY-MM-DD" eg:"2025-01-01"`
	EndTime   string `json:"endTime" dc:"创建时间范围-结束时间，格式YYYY-MM-DD" eg:"2025-12-31"`
}

// ListRes is the config list response.
type ListRes struct {
	List  []*entity.SysConfig `json:"list" dc:"参数设置列表" eg:"[]"`
	Total int                 `json:"total" dc:"总条数" eg:"10"`
}
