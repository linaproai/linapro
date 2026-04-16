// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysMenu is the golang structure of table sys_menu for DAO operations like Where/Data.
type SysMenu struct {
	g.Meta     `orm:"table:sys_menu, do:true"`
	Id         any         // 菜单ID
	ParentId   any         // 父菜单ID（0=根菜单）
	MenuKey    any         // 菜单稳定业务标识
	Name       any         // 菜单名称（支持i18n）
	Path       any         // 路由地址
	Component  any         // 组件路径
	Perms      any         // 权限标识
	Icon       any         // 菜单图标
	Type       any         // 菜单类型（D=目录 M=菜单 B=按钮）
	Sort       any         // 显示排序
	Visible    any         // 是否显示（1=显示 0=隐藏）
	Status     any         // 状态（0=停用 1=正常）
	IsFrame    any         // 是否外链（1=是 0=否）
	IsCache    any         // 是否缓存（1=是 0=否）
	QueryParam any         // 路由参数（JSON格式）
	Remark     any         // 备注
	CreatedAt  *gtime.Time // 创建时间
	UpdatedAt  *gtime.Time // 更新时间
	DeletedAt  *gtime.Time // 删除时间
}
