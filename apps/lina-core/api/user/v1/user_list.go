package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// ListReq defines the request for querying the user list.
type ListReq struct {
	g.Meta         `path:"/user" method:"get" tags:"User Management" summary:"Get user list" dc:"Query the paginated user list, support filtering by user name, nickname, status, mobile phone number, gender, department, creation time and other conditions, support custom sorting" permission:"system:user:query"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	Username       string `json:"username" dc:"Filter by username (fuzzy match)" eg:"admin"`
	Nickname       string `json:"nickname" dc:"Filter by nickname (fuzzy match)" eg:"Administrator"`
	Status         *int   `json:"status" dc:"Filter by status: 1=normal 0=disabled" eg:"1"`
	Phone          string `json:"phone" dc:"Filter by mobile phone number (fuzzy matching)" eg:"138"`
	Sex            *int   `json:"sex" dc:"Filter by gender: 0=Unknown 1=Male 2=Female" eg:"1"`
	DeptId         *int   `json:"deptId" dc:"Filter by department ID (including sub-departments)" eg:"100"`
	BeginTime      string `json:"beginTime" dc:"Filter by creation start time" eg:"2025-01-01"`
	EndTime        string `json:"endTime" dc:"Filter by creation end time" eg:"2025-12-31"`
	OrderBy        string `json:"orderBy" dc:"Sorting fields: id, username, nickname, phone, email, status, created_at" eg:"id"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"Sorting direction: asc or desc" eg:"desc"`
}

// ListItem represents a single user in the user list.
type ListItem struct {
	*entity.SysUser
	DeptId    int      `json:"deptId" dc:"Department ID" eg:"100"`
	DeptName  string   `json:"deptName" dc:"Department name" eg:"Technology Department"`
	RoleIds   []int    `json:"roleIds" dc:"Role ID list" eg:"[1,2]"`
	RoleNames []string `json:"roleNames" dc:"Role name list" eg:"[\"Administrator\",\"Normal User\"]"`
}

// ListRes is the response structure for user list query.
type ListRes struct {
	List  []*ListItem `json:"list" dc:"User list" eg:"[]"`
	Total int         `json:"total" dc:"Total number of items" eg:"100"`
}
