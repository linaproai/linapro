package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleDeleteReq is the request structure for role deletion.
type RoleDeleteReq struct {
	g.Meta `path:"/role/{id}" method:"delete" summary:"删除角色" tags:"角色管理" dc:"删除角色，会同时清理角色与菜单的关联关系、角色与用户的关联关系" permission:"system:role:remove"`
	Id     int `json:"id" v:"required|min:1" dc:"角色ID" eg:"3"`
}

// RoleDeleteRes is the response structure for role deletion.
type RoleDeleteRes struct {
	g.Meta `mime:"application/json" example:"{}"`
}
