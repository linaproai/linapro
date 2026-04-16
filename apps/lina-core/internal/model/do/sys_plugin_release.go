// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPluginRelease is the golang structure of table sys_plugin_release for DAO operations like Where/Data.
type SysPluginRelease struct {
	g.Meta           `orm:"table:sys_plugin_release, do:true"`
	Id               any         // 主键ID
	PluginId         any         // 插件唯一标识（kebab-case）
	ReleaseVersion   any         // 插件版本号
	Type             any         // 插件一级类型（source/dynamic）
	RuntimeKind      any         // 运行时产物类型（当前仅 wasm）
	SchemaVersion    any         // plugin.yaml 清单 schema 版本
	MinHostVersion   any         // 宿主最小兼容版本
	MaxHostVersion   any         // 宿主最大兼容版本
	Status           any         // release 状态（prepared/installed/active/uninstalled/failed）
	ManifestPath     any         // 插件清单路径
	PackagePath      any         // 插件源码目录或运行时产物路径
	Checksum         any         // 插件清单或产物校验值
	ManifestSnapshot any         // 插件清单与资源摘要快照（YAML，不保存具体 SQL/前端文件路径）
	CreatedAt        *gtime.Time // 创建时间
	UpdatedAt        *gtime.Time // 更新时间
	DeletedAt        *gtime.Time // 删除时间
}
