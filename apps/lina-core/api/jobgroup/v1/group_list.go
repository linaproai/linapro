package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq defines the request for querying scheduled job groups.
type ListReq struct {
	g.Meta         `path:"/job-group" method:"get" tags:"Job Scheduling / Job Groups" summary:"Get group list" dc:"Paginated query task grouping list, supports filtering by coding and name keywords" permission:"system:jobgroup:list"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	Code           string `json:"code" dc:"Filter by group encoding (fuzzy matching)" eg:"default"`
	Name           string `json:"name" dc:"Filter by group name (fuzzy matching)" eg:"Default grouping"`
	OrderBy        string `json:"orderBy" dc:"Sorting fields: id, sort_order, code, name, created_at, updated_at" eg:"sort_order"`
	OrderDirection string `json:"orderDirection" d:"asc" dc:"Sorting direction: asc=ascending order desc=descending order" eg:"asc"`
}

// ListItem represents one scheduled job group row in the list response.
type ListItem struct {
	JobGroupItem
	JobCount int64 `json:"jobCount" dc:"The number of tasks in the current group" eg:"3"`
}

// ListRes defines the response for querying scheduled job groups.
type ListRes struct {
	List  []*ListItem `json:"list" dc:"grouped list" eg:"[]"`
	Total int         `json:"total" dc:"Total number of items" eg:"1"`
}
