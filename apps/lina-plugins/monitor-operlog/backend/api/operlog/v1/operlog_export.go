package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// OperLog Export API

// ExportReq defines the request for exporting operation logs.
type ExportReq struct {
	g.Meta         `path:"/operlog/export" method:"get" tags:"Operation Logs" summary:"Export operation log" operLog:"export" dc:"Export operation logs to an Excel file, with optional filters or explicit record IDs." permission:"monitor:operlog:export"`
	Title          string  `json:"title" dc:"Filter by Module title (fuzzy match)" eg:"User Management"`
	OperName       string  `json:"operName" dc:"Filter by Operator (fuzzy match)" eg:"admin"`
	OperType       *string `json:"operType" v:"in:create,update,delete,export,import,other" dc:"Filter by operation type: create=new update=modify delete=delete export=export import=import other=other" eg:"create"`
	Status         *int    `json:"status" dc:"Filter by status: 0=success 1=failure" eg:"0"`
	BeginTime      string  `json:"beginTime" dc:"Filter by operation start time" eg:"2025-01-01"`
	EndTime        string  `json:"endTime" dc:"Filter by operation end time" eg:"2025-12-31"`
	OrderBy        string  `json:"orderBy" dc:"Sort fields: id,operTime,costTime" eg:"operTime"`
	OrderDirection string  `json:"orderDirection" d:"desc" dc:"Sort direction: asc or desc" eg:"desc"`
	Ids            []int   `json:"ids" dc:"Specify the list of record IDs to be exported. If omitted, all records that meet the conditions will be exported." eg:"[1,2,3]"`
}

// ExportRes is the operation-log export response.
type ExportRes struct{}
