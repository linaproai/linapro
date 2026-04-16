package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleUnassignUserReq is the request structure for unassigning user from role.
type RoleUnassignUserReq struct {
	g.Meta `path:"/role/{id}/users/{userId}" method:"delete" summary:"取消用户授权" tags:"角色管理" dc:"取消指定用户的角色授权，解除用户与角色的关联关系" permission:"system:role:auth"`
	Id     int `json:"id" v:"required|min:1" dc:"角色ID" eg:"1"`
	UserId int `json:"userId" v:"required|min:1" dc:"用户ID" eg:"2"`
}

// RoleUnassignUserRes is the response structure for unassigning user from role.
type RoleUnassignUserRes struct {
	g.Meta `mime:"application/json" example:"{}"`
}
