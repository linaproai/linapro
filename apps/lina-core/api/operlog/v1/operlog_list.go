package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// OperLog List API

// ListReq defines the request for listing operation logs.
type ListReq struct {
	g.Meta         `path:"/operlog" method:"get" tags:"操作日志" summary:"获取操作日志列表" dc:"分页查询操作日志列表，记录用户在系统中执行的增删改查等操作，支持多条件筛选和排序" permission:"monitor:operlog:query"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Title          string `json:"title" dc:"按模块标题筛选（模糊匹配）" eg:"用户管理"`
	OperName       string `json:"operName" dc:"按操作人员筛选（模糊匹配）" eg:"admin"`
	OperType       *int   `json:"operType" dc:"按操作类型筛选：1=新增 2=修改 3=删除 4=导出 5=导入" eg:"1"`
	Status         *int   `json:"status" dc:"按状态筛选：1=成功 0=失败" eg:"1"`
	BeginTime      string `json:"beginTime" dc:"按操作时间起始筛选" eg:"2025-01-01"`
	EndTime        string `json:"endTime" dc:"按操作时间结束筛选" eg:"2025-12-31"`
	OrderBy        string `json:"orderBy" dc:"排序字段：id,oper_time,cost_time" eg:"oper_time"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"排序方向：asc或desc" eg:"desc"`
}

// ListRes Operation log list response
type ListRes struct {
	Items []*entity.SysOperLog `json:"items" dc:"操作日志列表" eg:"[]"`
	Total int                  `json:"total" dc:"总条数" eg:"500"`
}
