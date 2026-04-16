// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysDept is the golang structure of table sys_dept for DAO operations like Where/Data.
type SysDept struct {
	g.Meta    `orm:"table:sys_dept, do:true"`
	Id        any         // 部门ID
	ParentId  any         // 父部门ID
	Ancestors any         // 祖级列表
	Name      any         // 部门名称
	Code      any         // 部门编码
	OrderNum  any         // 显示排序
	Leader    any         // 负责人用户ID
	Phone     any         // 联系电话
	Email     any         // 邮箱
	Status    any         // 状态（0停用 1正常）
	Remark    any         // 备注
	CreatedAt *gtime.Time // 创建时间
	UpdatedAt *gtime.Time // 更新时间
	DeletedAt *gtime.Time // 删除时间
}
