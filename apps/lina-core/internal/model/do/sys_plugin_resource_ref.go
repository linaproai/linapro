// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPluginResourceRef is the golang structure of table sys_plugin_resource_ref for DAO operations like Where/Data.
type SysPluginResourceRef struct {
	g.Meta       `orm:"table:sys_plugin_resource_ref, do:true"`
	Id           any         // 主键ID
	PluginId     any         // 插件唯一标识（kebab-case）
	ReleaseId    any         // 所属插件 release ID
	ResourceType any         // 资源类型（manifest/sql/frontend/menu/permission 等）
	ResourceKey  any         // 资源唯一键
	ResourcePath any         // 资源定位补充信息（默认留空，不保存具体前端/SQL 路径）
	OwnerType    any         // 宿主对象类型（file/menu/route/slot 等）
	OwnerKey     any         // 宿主对象稳定标识
	Remark       any         // 备注
	CreatedAt    *gtime.Time // 创建时间
	UpdatedAt    *gtime.Time // 更新时间
	DeletedAt    *gtime.Time // 删除时间
}
