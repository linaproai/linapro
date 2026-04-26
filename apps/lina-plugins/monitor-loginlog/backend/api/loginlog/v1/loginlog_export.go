package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// LoginLog Export API

// ExportReq defines the request for exporting login logs.
type ExportReq struct {
	g.Meta         `path:"/loginlog/export" method:"get" tags:"Login Logs" summary:"Export login logs" operLog:"export" dc:"Export login logs to an Excel file, with optional filters or explicit record IDs." permission:"monitor:loginlog:export"`
	UserName       string `json:"userName" dc:"Filter by username (fuzzy match)" eg:"admin"`
	Ip             string `json:"ip" dc:"Filter by IP address (fuzzy match)" eg:"192.168"`
	Status         *int   `json:"status" dc:"Filter by status: 0=success 1=failure" eg:"0"`
	BeginTime      string `json:"beginTime" dc:"Filter by login start time" eg:"2025-01-01"`
	EndTime        string `json:"endTime" dc:"Filter by login end time" eg:"2025-12-31"`
	OrderBy        string `json:"orderBy" dc:"Sort fields: id, loginTime" eg:"loginTime"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"Sort direction: asc or desc" eg:"desc"`
	Ids            []int  `json:"ids" dc:"Specify the list of record IDs to be exported. If omitted, all records that meet the conditions will be exported." eg:"[1,2,3]"`
}

// ExportRes is the login-log export response.
type ExportRes struct{}
