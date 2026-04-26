package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleStatusReq is the request structure for role status toggle.
type RoleStatusReq struct {
	g.Meta `path:"/role/{id}/status" method:"put" summary:"Switch role status" tags:"Role Management" dc:"Switch role status (enable/disable). Users with deactivated roles will not be able to use their permissions." permission:"system:role:edit"`
	Id     int `json:"id" v:"required|min:1" dc:"Role ID" eg:"1"`
	Status int `json:"status" v:"required|in:0,1" dc:"Target status: 1=normal 0=disabled" eg:"0"`
}

// RoleStatusRes is the response structure for role status toggle.
type RoleStatusRes struct {
	g.Meta `mime:"application/json" example:"{}"`
}
