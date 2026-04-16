package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleStatusReq is the request structure for role status toggle.
type RoleStatusReq struct {
	g.Meta `path:"/role/{id}/status" method:"put" summary:"切换角色状态" tags:"角色管理" dc:"切换角色状态（启用/停用），停用的角色用户将无法使用其权限" permission:"system:role:edit"`
	Id     int `json:"id" v:"required|min:1" dc:"角色ID" eg:"1"`
	Status int `json:"status" v:"required|in:0,1" dc:"目标状态：1=正常 0=停用" eg:"0"`
}

// RoleStatusRes is the response structure for role status toggle.
type RoleStatusRes struct {
	g.Meta `mime:"application/json" example:"{}"`
}
