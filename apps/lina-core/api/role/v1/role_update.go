package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleUpdateReq is the request structure for role update.
type RoleUpdateReq struct {
	g.Meta    `path:"/role/{id}" method:"put" summary:"Update role" tags:"Role Management" dc:"Update role information. You can modify the role name, permission key, status, menu association, etc. The name and permission key cannot be repeated with other roles." permission:"system:role:edit"`
	Id        int    `json:"id" v:"required|min:1" dc:"Role ID" eg:"1"`
	Name      string `json:"name" v:"required|length:2,30" dc:"Role name, length 2-30 characters" eg:"Administrator"`
	Key       string `json:"key" v:"required|length:2,30" dc:"Permission key, length 2-30 characters" eg:"admin"`
	Sort      int    `json:"sort" v:"min:0" dc:"Show sort" eg:"1"`
	DataScope int    `json:"dataScope" v:"in:1,2,3" dc:"Data permission scope: 1=all 2=this department 3=only me" eg:"1"`
	Status    int    `json:"status" v:"in:0,1" dc:"Status: 1=normal 0=disabled" eg:"1"`
	Remark    string `json:"remark" v:"length:0,200" dc:"Remarks" eg:"system administrator"`
	MenuIds   []int  `json:"menuIds" dc:"List of associated menu IDs" eg:"[1,2,3,10]"`
}

// RoleUpdateRes is the response structure for role update.
type RoleUpdateRes struct {
	g.Meta `mime:"application/json" example:"{}"`
}
