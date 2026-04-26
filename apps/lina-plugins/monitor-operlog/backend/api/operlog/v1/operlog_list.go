package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// OperLog List API

// ListReq defines the request for listing operation logs.
type ListReq struct {
	g.Meta         `path:"/operlog" method:"get" tags:"Operation Logs" summary:"Get operation log list" dc:"Query operation logs by page, recording create, update, delete, and other user operations with multi-condition filtering and sorting." permission:"monitor:operlog:query"`
	PageNum        int     `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize       int     `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	Title          string  `json:"title" dc:"Filter by Module title (fuzzy match)" eg:"User Management"`
	OperName       string  `json:"operName" dc:"Filter by Operator (fuzzy match)" eg:"admin"`
	OperType       *string `json:"operType" v:"in:create,update,delete,export,import,other" dc:"Filter by operation type: create=new update=modify delete=delete export=export import=import other=other" eg:"create"`
	Status         *int    `json:"status" dc:"Filter by status: 0=success 1=failure" eg:"0"`
	BeginTime      string  `json:"beginTime" dc:"Filter by operation start time" eg:"2025-01-01"`
	EndTime        string  `json:"endTime" dc:"Filter by operation end time" eg:"2025-12-31"`
	OrderBy        string  `json:"orderBy" dc:"Sort fields: id,operTime,costTime" eg:"operTime"`
	OrderDirection string  `json:"orderDirection" d:"desc" dc:"Sort direction: asc or desc" eg:"desc"`
}

// ListRes is the operation-log list response.
type ListRes struct {
	Items []*OperLogEntity `json:"items" dc:"Operation log list" eg:"[]"`
	Total int              `json:"total" dc:"Total number of items" eg:"500"`
}
