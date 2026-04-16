// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPluginState is the golang structure of table sys_plugin_state for DAO operations like Where/Data.
type SysPluginState struct {
	g.Meta     `orm:"table:sys_plugin_state, do:true"`
	Id         any         // 主键ID
	PluginId   any         // 插件唯一标识（kebab-case）
	StateKey   any         // 状态键
	StateValue any         // 状态值（支持JSON）
	CreatedAt  *gtime.Time // 创建时间
	UpdatedAt  *gtime.Time // 更新时间
}
