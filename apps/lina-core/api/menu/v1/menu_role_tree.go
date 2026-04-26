package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleMenuTreeReq defines the request for querying a role's menu tree.
type RoleMenuTreeReq struct {
	g.Meta `path:"/menu/role/{roleId}" method:"get" tags:"Menu Management" summary:"Get the role menu tree" dc:"Get the role menu tree, used to display assigned menus when editing a role. Returns all menu trees and the menu IDs assigned to this role" permission:"system:menu:query"`
	RoleId int `json:"roleId" v:"required|min:1" dc:"Role ID" eg:"1"`
}

// RoleMenuTreeRes defines the response for querying a role's menu tree.
type RoleMenuTreeRes struct {
	Menus       []*MenuTreeNode `json:"menus" dc:"Menu tree list" eg:"[]"`
	CheckedKeys []int           `json:"checkedKeys" dc:"Checked menu ID list" eg:"[1,2,3]"`
}
