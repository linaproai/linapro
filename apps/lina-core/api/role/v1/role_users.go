package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleUsersReq is the request structure for role user list query.
type RoleUsersReq struct {
	g.Meta   `path:"/role/{id}/users" method:"get" summary:"Query role user list" tags:"Role Management" dc:"Query the paginated list of users assigned to the specified role, and support filtering by user name, mobile phone number, etc." permission:"system:role:auth"`
	Id       int    `json:"id" v:"required|min:1" dc:"Role ID" eg:"1"`
	Username string `json:"username" dc:"Username, fuzzy query" eg:"admin"`
	Phone    string `json:"phone" dc:"Mobile phone number, fuzzy query" eg:"138"`
	Status   int    `json:"status" dc:"Status filtering: 1=normal 0=disabled, if not transmitted, all will be queried" eg:"1"`
	Page     int    `json:"page" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	Size     int    `json:"size" d:"10" v:"min:1|max:100" dc:"Number of records per page" eg:"10"`
}

// RoleUsersRes is the response structure for role user list query.
type RoleUsersRes struct {
	g.Meta `mime:"application/json" example:"{}"`
	List   []*RoleUserItem `json:"list" dc:"User list" eg:"[]"`
	Total  int             `json:"total" dc:"Total number of records" eg:"10"`
}

// RoleUserItem represents a single user in the role user list.
type RoleUserItem struct {
	Id        int    `json:"id" dc:"User ID" eg:"1"`
	Username  string `json:"username" dc:"Username" eg:"admin"`
	Nickname  string `json:"nickname" dc:"Nickname" eg:"Administrator"`
	Email     string `json:"email" dc:"Email" eg:"admin@example.com"`
	Phone     string `json:"phone" dc:"Mobile phone number" eg:"13800138000"`
	Status    int    `json:"status" dc:"Status (0=disabled 1=normal)" eg:"1"`
	CreatedAt string `json:"createdAt" dc:"creation time" eg:"2024-01-01 00:00:00"`
}
