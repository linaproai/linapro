package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq defines the request for querying the department list.
type ListReq struct {
	g.Meta `path:"/dept" method:"get" tags:"Department Management" summary:"Get department list" dc:"Get departments at all levels, with filtering by department name and status. Results are sorted by sort order in ascending order." permission:"system:dept:query"`
	Name   string `json:"name" dc:"Filter by department name, support fuzzy matching" eg:"technology"`
	Status *int   `json:"status" dc:"Filter by status: 1=normal, 0=disabled, if omitted, query all" eg:"1"`
}

// ListRes defines the response for querying the department list.
type ListRes struct {
	List []*DeptEntity `json:"list" dc:"Department list data, including all department records matching the conditions" eg:"[]"`
}
