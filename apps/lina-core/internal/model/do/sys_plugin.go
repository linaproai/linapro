// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPlugin is the golang structure of table sys_plugin for DAO operations like Where/Data.
type SysPlugin struct {
	g.Meta       `orm:"table:sys_plugin, do:true"`
	Id           any         // 主键ID
	PluginId     any         // 插件唯一标识（kebab-case）
	Name         any         // 插件名称
	Version      any         // 插件版本号
	Type         any         // 插件一级类型（source/dynamic）
	Installed    any         // 安装状态（1=已安装 0=未安装）
	Status       any         // 启用状态（1=启用 0=禁用）
	DesiredState any         // 宿主期望状态（uninstalled/installed/enabled）
	CurrentState any         // 宿主当前状态（uninstalled/installed/enabled/reconciling/failed）
	Generation   any         // 宿主当前生效代际号
	ReleaseId    any         // 宿主当前生效 release ID
	ManifestPath any         // 插件清单文件路径
	Checksum     any         // 插件包校验值
	InstalledAt  *gtime.Time // 安装时间
	EnabledAt    *gtime.Time // 最后一次启用时间
	DisabledAt   *gtime.Time // 最后一次禁用时间
	Remark       any         // 备注
	CreatedAt    *gtime.Time // 创建时间
	UpdatedAt    *gtime.Time // 更新时间
	DeletedAt    *gtime.Time // 删除时间
}
