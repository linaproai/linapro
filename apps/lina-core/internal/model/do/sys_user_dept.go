// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SysUserDept is the golang structure of table sys_user_dept for DAO operations like Where/Data.
type SysUserDept struct {
	g.Meta `orm:"table:sys_user_dept, do:true"`
	UserId any // 用户ID
	DeptId any // 部门ID
}
