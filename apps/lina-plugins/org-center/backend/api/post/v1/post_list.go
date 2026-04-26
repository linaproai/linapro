package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq defines the request for querying the post list.
type ListReq struct {
	g.Meta   `path:"/post" method:"get" tags:"Position Management" summary:"Get position list" dc:"Query the position list by page, with filtering by department, code, name, and status." permission:"system:post:query"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	DeptId   *int   `json:"deptId" dc:"Filter by department ID" eg:"100"`
	Code     string `json:"code" dc:"Filter by position code (fuzzy matching)" eg:"ceo"`
	Name     string `json:"name" dc:"Filter by job title (fuzzy matching)" eg:"general manager"`
	Status   *int   `json:"status" dc:"Filter by status: 1=normal 0=disabled" eg:"1"`
}

// ListRes is the response for post list.
type ListRes struct {
	List  []*PostEntity `json:"list" dc:"Job list" eg:"[]"`
	Total int           `json:"total" dc:"Total number of items" eg:"20"`
}
