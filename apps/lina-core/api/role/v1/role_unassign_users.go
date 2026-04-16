package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleUnassignUsersReq is the request structure for batch unassigning users from role.
type RoleUnassignUsersReq struct {
	g.Meta  `path:"/role/{id}/users" method:"delete" summary:"批量取消用户授权" tags:"角色管理" dc:"批量取消多个用户的角色授权，解除用户与角色的关联关系" permission:"system:role:auth"`
	Id      int   `json:"id" v:"required|min:1" dc:"角色ID" eg:"1"`
	UserIds []int `json:"userIds" v:"required|min:1" dc:"用户ID列表" eg:"[1,2,3]"`
}

// RoleUnassignUsersRes is the response structure for batch unassigning users from role.
type RoleUnassignUsersRes struct {
	g.Meta `mime:"application/json" example:"{}"`
}
