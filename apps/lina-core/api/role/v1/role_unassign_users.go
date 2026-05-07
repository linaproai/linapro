package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleUnassignUsersReq is the request structure for batch unassigning users from role.
type RoleUnassignUsersReq struct {
	g.Meta  `path:"/role/{id}/users" method:"delete" summary:"Cancel user authorization in batches" tags:"Role Management" dc:"Cancel the role authorization of multiple users in batches and remove the association between users and roles." permission:"system:role:auth"`
	Id      int   `json:"id" v:"required|min:1" dc:"Role ID" eg:"1"`
	UserIds []int `json:"userIds" v:"required|min-length:1" dc:"User ID list" eg:"[1,2,3]"`
}

// RoleUnassignUsersRes is the response structure for batch unassigning users from role.
type RoleUnassignUsersRes struct {
	g.Meta `mime:"application/json" example:"{}"`
}
