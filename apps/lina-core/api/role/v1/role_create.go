package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleCreateReq is the request structure for role creation.
type RoleCreateReq struct {
	g.Meta    `path:"/role" method:"post" summary:"Create a role" tags:"Role Management" dc:"Create a new role. The role name and permission key must be unique and can be associated with a menu." permission:"system:role:add"`
	Name      string `json:"name" v:"required|length:2,30" dc:"Role name, length 2-30 characters" eg:"test role"`
	Key       string `json:"key" v:"required|length:2,30" dc:"Permission key, length 2-30 characters, used for permission identification" eg:"test_role"`
	Sort      int    `json:"sort" d:"0" v:"min:0" dc:"Display sorting, the smaller the number, the higher it is" eg:"0"`
	DataScope int    `json:"dataScope" d:"2" v:"in:1,2,3,4" dc:"Data permission scope: 1=all 2=current tenant 3=current department 4=only me" eg:"2"`
	Status    int    `json:"status" d:"1" v:"in:0,1" dc:"Status: 1=normal 0=disabled" eg:"1"`
	Remark    string `json:"remark" v:"length:0,200" dc:"Remarks, up to 200 characters" eg:"Test role description"`
	MenuIds   []int  `json:"menuIds" dc:"A list of associated menu IDs used to control the role's menu permissions" eg:"[1,2,3]"`
}

// RoleCreateRes is the response structure for role creation.
type RoleCreateRes struct {
	g.Meta `mime:"application/json" example:"{}"`
	Id     int `json:"id" dc:"Role ID after successful creation" eg:"3"`
}
