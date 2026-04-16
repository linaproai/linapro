package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleAssignUsersReq is the request structure for assigning users to role.
type RoleAssignUsersReq struct {
	g.Meta  `path:"/role/{id}/users" method:"post" summary:"分配用户到角色" tags:"角色管理" dc:"批量分配用户到指定角色，已分配的用户不会被重复添加" permission:"system:role:auth"`
	Id      int   `json:"id" v:"required|min:1" dc:"角色ID" eg:"1"`
	UserIds []int `json:"userIds" v:"required|min:1" dc:"要分配的用户ID列表" eg:"[2,3,4]"`
}

// RoleAssignUsersRes is the response structure for assigning users to role.
type RoleAssignUsersRes struct {
	g.Meta `mime:"application/json" example:"{}"`
}
