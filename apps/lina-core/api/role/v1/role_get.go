package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleGetReq is the request structure for role detail query.
type RoleGetReq struct {
	g.Meta `path:"/role/{id}" method:"get" summary:"Query role details" tags:"Role Management" dc:"Query role details based on role ID, including associated menu ID list" permission:"system:role:query"`
	Id     int `json:"id" v:"required|min:1" dc:"Role ID" eg:"1"`
}

// RoleGetRes is the response structure for role detail query.
type RoleGetRes struct {
	g.Meta    `mime:"application/json" example:"{}"`
	Id        int    `json:"id" dc:"Role ID" eg:"1"`
	Name      string `json:"name" dc:"Character name" eg:"Administrator"`
	Key       string `json:"key" dc:"permission key" eg:"admin"`
	Sort      int    `json:"sort" dc:"Show sort" eg:"1"`
	DataScope int    `json:"dataScope" dc:"Data permission scope (1=all 2=current tenant 3=current department 4=only me)" eg:"2"`
	Status    int    `json:"status" dc:"Status (0=disabled 1=normal)" eg:"1"`
	Remark    string `json:"remark" dc:"Remarks" eg:"System administrator role"`
	MenuIds   []int  `json:"menuIds" dc:"List of associated menu IDs" eg:"[1,2,3]"`
	CreatedAt string `json:"createdAt" dc:"creation time" eg:"2024-01-01 00:00:00"`
	UpdatedAt string `json:"updatedAt" dc:"Update time" eg:"2024-01-01 00:00:00"`
}
