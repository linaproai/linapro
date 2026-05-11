package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleListReq is the request structure for role list query.
type RoleListReq struct {
	g.Meta `path:"/role" method:"get" summary:"Query role list" tags:"Role Management" dc:"Query the paginated role list, supporting filtering by role name, permission key, status, etc." permission:"system:role:query"`
	Name   string `json:"name" dc:"Role name, fuzzy query" eg:"Administrator"`
	Key    string `json:"key" dc:"Permission key, fuzzy query" eg:"admin"`
	Status int    `json:"status" dc:"Status filtering: 1=normal 0=disabled, if not transmitted, all will be queried" eg:"1"`
	Page   int    `json:"page" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	Size   int    `json:"size" d:"10" v:"min:1|max:100" dc:"Number of records per page" eg:"10"`
}

// RoleListRes is the response structure for role list query.
type RoleListRes struct {
	g.Meta `mime:"application/json" example:"{}"`
	List   []*RoleListItem `json:"list" dc:"role list" eg:"[]"`
	Total  int             `json:"total" dc:"Total number of records" eg:"10"`
}

// RoleListItem represents a single role in the list.
type RoleListItem struct {
	Id        int    `json:"id" dc:"Role ID" eg:"1"`
	Name      string `json:"name" dc:"Character name" eg:"Administrator"`
	Key       string `json:"key" dc:"permission key" eg:"admin"`
	Sort      int    `json:"sort" dc:"Show sort" eg:"1"`
	DataScope int    `json:"dataScope" dc:"Data permission scope (1=all 2=current tenant 3=current department 4=only me)" eg:"2"`
	Status    int    `json:"status" dc:"Status (0=disabled 1=normal)" eg:"1"`
	Remark    string `json:"remark" dc:"Remarks" eg:"System administrator role"`
	CreatedAt string `json:"createdAt" dc:"creation time" eg:"2024-01-01 00:00:00"`
	UpdatedAt string `json:"updatedAt" dc:"Update time" eg:"2024-01-01 00:00:00"`
}
