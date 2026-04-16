// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysRole is the golang structure of table sys_role for DAO operations like Where/Data.
type SysRole struct {
	g.Meta    `orm:"table:sys_role, do:true"`
	Id        any         // 角色ID
	Name      any         // 角色名称
	Key       any         // 权限字符
	Sort      any         // 显示排序
	DataScope any         // 数据权限范围（1=全部 2=本部门 3=仅本人）
	Status    any         // 状态（0=停用 1=正常）
	Remark    any         // 备注
	CreatedAt *gtime.Time // 创建时间
	UpdatedAt *gtime.Time // 更新时间
	DeletedAt *gtime.Time // 删除时间
}
