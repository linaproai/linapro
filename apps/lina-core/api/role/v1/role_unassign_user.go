package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleUnassignUserReq is the request structure for unassigning user from role.
type RoleUnassignUserReq struct {
	g.Meta `path:"/role/{id}/users/{userId}" method:"delete" summary:"Cancel user authorization" tags:"Role Management" dc:"Cancel the role authorization of the specified user and remove the association between the user and the role" permission:"system:role:auth"`
	Id     int `json:"id" v:"required|min:1" dc:"Role ID" eg:"1"`
	UserId int `json:"userId" v:"required|min:1" dc:"User ID" eg:"2"`
}

// RoleUnassignUserRes is the response structure for unassigning user from role.
type RoleUnassignUserRes struct {
	g.Meta `mime:"application/json" example:"{}"`
}
