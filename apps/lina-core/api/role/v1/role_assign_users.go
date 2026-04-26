package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleAssignUsersReq is the request structure for assigning users to role.
type RoleAssignUsersReq struct {
	g.Meta  `path:"/role/{id}/users" method:"post" summary:"Assign users to roles" tags:"Role Management" dc:"Assign users to specified roles in batches. Already assigned users will not be added repeatedly." permission:"system:role:auth"`
	Id      int   `json:"id" v:"required|min:1" dc:"Role ID" eg:"1"`
	UserIds []int `json:"userIds" v:"required|min:1" dc:"List of user IDs to assign" eg:"[2,3,4]"`
}

// RoleAssignUsersRes is the response structure for assigning users to role.
type RoleAssignUsersRes struct {
	g.Meta `mime:"application/json" example:"{}"`
}
