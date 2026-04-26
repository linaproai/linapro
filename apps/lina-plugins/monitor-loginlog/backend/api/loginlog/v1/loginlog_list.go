package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// LoginLog List API

// ListReq defines the request for listing login logs.
type ListReq struct {
	g.Meta         `path:"/loginlog" method:"get" tags:"Login Logs" summary:"Get login log list" dc:"Query login logs by page, recording successful and failed logins with multi-condition filtering and sorting." permission:"monitor:loginlog:query"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	UserName       string `json:"userName" dc:"Filter by username (fuzzy match)" eg:"admin"`
	Ip             string `json:"ip" dc:"Filter by IP address (fuzzy match)" eg:"192.168"`
	Status         *int   `json:"status" dc:"Filter by status: 0=success 1=failure" eg:"0"`
	BeginTime      string `json:"beginTime" dc:"Filter by login start time" eg:"2025-01-01"`
	EndTime        string `json:"endTime" dc:"Filter by login end time" eg:"2025-12-31"`
	OrderBy        string `json:"orderBy" dc:"Sort fields: id, loginTime" eg:"loginTime"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"Sort direction: asc or desc" eg:"desc"`
}

// ListRes is the login-log list response.
type ListRes struct {
	Items []*LoginLogEntity `json:"items" dc:"Login log list" eg:"[]"`
	Total int               `json:"total" dc:"Total number of items" eg:"100"`
}
