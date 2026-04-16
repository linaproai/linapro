package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ExportReq defines the request for exporting posts.
type ExportReq struct {
	g.Meta `path:"/post/export" method:"get" tags:"岗位管理" summary:"导出岗位数据" operLog:"export" dc:"导出岗位数据为Excel文件，支持按条件筛选导出" permission:"system:post:export"`
	DeptId *int   `json:"deptId" dc:"按部门ID筛选" eg:"100"`
	Code   string `json:"code" dc:"按岗位编码筛选" eg:"dev"`
	Name   string `json:"name" dc:"按岗位名称筛选" eg:"工程师"`
	Status *int   `json:"status" dc:"按状态筛选：1=正常 0=停用" eg:"1"`
}

// ExportRes defines the response for exporting posts.
type ExportRes struct{}
