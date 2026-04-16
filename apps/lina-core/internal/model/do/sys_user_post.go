// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SysUserPost is the golang structure of table sys_user_post for DAO operations like Where/Data.
type SysUserPost struct {
	g.Meta `orm:"table:sys_user_post, do:true"`
	UserId any // 用户ID
	PostId any // 岗位ID
}
