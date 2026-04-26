package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleDeleteReq is the request structure for role deletion.
type RoleDeleteReq struct {
	g.Meta `path:"/role/{id}" method:"delete" summary:"Delete role" tags:"Role Management" dc:"Deleting a role will also clear the relationship between the role and the menu, and the relationship between the role and the user." permission:"system:role:remove"`
	Id     int `json:"id" v:"required|min:1" dc:"Role ID" eg:"3"`
}

// RoleDeleteRes is the response structure for role deletion.
type RoleDeleteRes struct {
	g.Meta `mime:"application/json" example:"{}"`
}
