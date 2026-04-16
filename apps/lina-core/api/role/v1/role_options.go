package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleOptionsReq is the request structure for role options query.
type RoleOptionsReq struct {
	g.Meta `path:"/role/options" method:"get" summary:"查询角色下拉选项" tags:"角色管理" dc:"查询所有启用状态的角色下拉选项列表，用于用户角色选择等场景" permission:"system:role:query"`
}

// RoleOptionsRes is the response structure for role options query.
type RoleOptionsRes struct {
	g.Meta `mime:"application/json" example:"{}"`
	List   []*RoleOptionItem `json:"list" dc:"角色下拉选项列表" eg:"[]"`
}

// RoleOptionItem represents a single role option.
type RoleOptionItem struct {
	Id   int    `json:"id" dc:"角色ID" eg:"1"`
	Name string `json:"name" dc:"角色名称" eg:"管理员"`
	Key  string `json:"key" dc:"权限字符" eg:"admin"`
}
