package v1

import "github.com/gogf/gf/v2/frame/g"

// Config List API

// ListReq defines the request for querying config list.
type ListReq struct {
	g.Meta    `path:"/config" method:"get" tags:"Parameter Settings" summary:"Get parameter setting list" dc:"Paginated query parameter setting list, supports fuzzy matching by parameter name, parameter key name, and filtering by creation time range" permission:"system:config:query"`
	PageNum   int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize  int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	Name      string `json:"name" dc:"Filter by parameter name (fuzzy matching), query all if not passed" eg:"Main frame page"`
	Key       string `json:"key" dc:"Filter by parameter key name (fuzzy matching), query all if not passed" eg:"sys.index"`
	BeginTime string `json:"beginTime" dc:"Creation time range-start time, format YYYY-MM-DD" eg:"2025-01-01"`
	EndTime   string `json:"endTime" dc:"Creation time range-end time, format YYYY-MM-DD" eg:"2025-12-31"`
}

// ListRes is the config list response.
type ListRes struct {
	List  []*ConfigItem `json:"list" dc:"Parameter setting list" eg:"[]"`
	Total int           `json:"total" dc:"Total number of items" eg:"10"`
}
