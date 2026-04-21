// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysUser is the golang structure of table sys_user for DAO operations like Where/Data.
type SysUser struct {
	g.Meta    `orm:"table:sys_user, do:true"`
	Id        any         // 用户ID
	Username  any         // 用户账号
	Password  any         // 密码
	Nickname  any         // 用户昵称
	Email     any         // 邮箱
	Phone     any         // 手机号码
	Sex       any         // 性别（0未知 1男 2女）
	Avatar    any         // 头像地址
	Status    any         // 状态（0停用 1正常）
	Remark    any         // 备注
	LoginDate *gtime.Time // 最后登录时间
	CreatedAt *gtime.Time // 创建时间
	UpdatedAt *gtime.Time // 更新时间
	DeletedAt *gtime.Time // 删除时间
}
