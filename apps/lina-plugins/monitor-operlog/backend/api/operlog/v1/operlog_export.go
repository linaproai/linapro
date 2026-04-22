package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// OperLog Export API

// ExportReq defines the request for exporting operation logs.
type ExportReq struct {
	g.Meta         `path:"/operlog/export" method:"get" tags:"操作日志" summary:"导出操作日志" operLog:"export" dc:"导出操作日志数据为Excel文件，支持按条件筛选导出，也支持导出指定ID的记录" permission:"monitor:operlog:export"`
	Title          string  `json:"title" dc:"按模块标题筛选（模糊匹配）" eg:"用户管理"`
	OperName       string  `json:"operName" dc:"按操作人员筛选（模糊匹配）" eg:"admin"`
	OperType       *string `json:"operType" v:"in:create,update,delete,export,import,other" dc:"按操作类型筛选：create=新增 update=修改 delete=删除 export=导出 import=导入 other=其他" eg:"create"`
	Status         *int    `json:"status" dc:"按状态筛选：0=成功 1=失败" eg:"0"`
	BeginTime      string  `json:"beginTime" dc:"按操作时间起始筛选" eg:"2025-01-01"`
	EndTime        string  `json:"endTime" dc:"按操作时间结束筛选" eg:"2025-12-31"`
	OrderBy        string  `json:"orderBy" dc:"排序字段：id,operTime,costTime" eg:"operTime"`
	OrderDirection string  `json:"orderDirection" d:"desc" dc:"排序方向：asc或desc" eg:"desc"`
	Ids            []int   `json:"ids" dc:"指定导出的记录ID列表，不传则导出全部符合条件的记录" eg:"[1,2,3]"`
}

// ExportRes is the operation-log export response.
type ExportRes struct{}
