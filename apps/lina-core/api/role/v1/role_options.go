package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleOptionsReq is the request structure for role options query.
type RoleOptionsReq struct {
	g.Meta `path:"/role/options" method:"get" summary:"Query role dropdown options" tags:"Role Management" dc:"Query the dropdown option list of all enabled roles for user role selection and other scenarios." permission:"system:role:query"`
}

// RoleOptionsRes is the response structure for role options query.
type RoleOptionsRes struct {
	g.Meta `mime:"application/json" example:"{}"`
	List   []*RoleOptionItem `json:"list" dc:"Role dropdown list" eg:"[]"`
}

// RoleOptionItem represents a single role option.
type RoleOptionItem struct {
	Id   int    `json:"id" dc:"Role ID" eg:"1"`
	Name string `json:"name" dc:"Character name" eg:"Administrator"`
	Key  string `json:"key" dc:"permission key" eg:"admin"`
}
