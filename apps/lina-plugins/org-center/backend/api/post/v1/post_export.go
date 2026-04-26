package v1

import "github.com/gogf/gf/v2/frame/g"

// ExportReq defines the request for exporting posts.
type ExportReq struct {
	g.Meta `path:"/post/export" method:"get" tags:"Position Management" summary:"Export position data" operLog:"export" dc:"Export position data to an Excel file, with optional filters." permission:"system:post:export"`
	DeptId *int   `json:"deptId" dc:"Filter by department ID" eg:"100"`
	Code   string `json:"code" dc:"Filter by job code" eg:"dev"`
	Name   string `json:"name" dc:"Filter by job title" eg:"Engineer"`
	Status *int   `json:"status" dc:"Filter by status: 1=normal 0=disabled" eg:"1"`
}

// ExportRes defines the response for exporting posts.
type ExportRes struct{}
