package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// OperLog Export API

// ExportReq defines the request for exporting operation logs.
type ExportReq struct {
	g.Meta         `path:"/operlog/export" method:"get" tags:"操作日志" summary:"导出操作日志" operLog:"export" dc:"导出操作日志数据为Excel文件，支持按条件筛选导出，也支持导出指定ID的记录" permission:"monitor:operlog:export"`
	Title          string `json:"title" dc:"按模块标题筛选（模糊匹配）" eg:"用户管理"`
	OperName       string `json:"operName" dc:"按操作人员筛选（模糊匹配）" eg:"admin"`
	OperType       *int   `json:"operType" dc:"按操作类型筛选：1=新增 2=修改 3=删除 4=导出 5=导入" eg:"1"`
	Status         *int   `json:"status" dc:"按状态筛选：1=成功 0=失败" eg:"1"`
	BeginTime      string `json:"beginTime" dc:"按操作时间起始筛选" eg:"2025-01-01"`
	EndTime        string `json:"endTime" dc:"按操作时间结束筛选" eg:"2025-12-31"`
	OrderBy        string `json:"orderBy" dc:"排序字段" eg:"oper_time"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"排序方向：asc或desc" eg:"desc"`
	Ids            []int  `json:"ids" dc:"指定导出的记录ID列表，不传则导出全部符合条件的记录" eg:"[1,2,3]"`
}

// ExportRes Operation log export response
type ExportRes struct{}
