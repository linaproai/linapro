package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleMenuTreeReq defines the request for querying a role's menu tree.
type RoleMenuTreeReq struct {
	g.Meta `path:"/menu/role/{roleId}" method:"get" tags:"菜单管理" summary:"获取角色的菜单树" dc:"获取角色的菜单树，用于角色编辑时显示已分配的菜单。返回所有菜单树和该角色已分配的菜单ID列表" permission:"system:menu:query"`
	RoleId int `json:"roleId" v:"required|min:1" dc:"角色ID" eg:"1"`
}

// RoleMenuTreeRes defines the response for querying a role's menu tree.
type RoleMenuTreeRes struct {
	Menus       []*MenuTreeNode `json:"menus" dc:"菜单树形列表" eg:"[]"`
	CheckedKeys []int           `json:"checkedKeys" dc:"已勾选的菜单ID列表" eg:"[1,2,3]"`
}
