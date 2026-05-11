// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SysUserRole is the golang structure of table sys_user_role for DAO operations like Where/Data.
type SysUserRole struct {
	g.Meta   `orm:"table:sys_user_role, do:true"`
	TenantId any // Role assignment tenant ID, 0 means PLATFORM
	UserId   any // User ID
	RoleId   any // Role ID
}
