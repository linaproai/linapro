// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SysRoleMenu is the golang structure of table sys_role_menu for DAO operations like Where/Data.
type SysRoleMenu struct {
	g.Meta `orm:"table:sys_role_menu, do:true"`
	RoleId any // 角色ID
	MenuId any // 菜单ID
}
