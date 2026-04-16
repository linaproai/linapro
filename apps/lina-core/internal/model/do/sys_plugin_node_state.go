// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPluginNodeState is the golang structure of table sys_plugin_node_state for DAO operations like Where/Data.
type SysPluginNodeState struct {
	g.Meta          `orm:"table:sys_plugin_node_state, do:true"`
	Id              any         // 主键ID
	PluginId        any         // 插件唯一标识（kebab-case）
	ReleaseId       any         // 所属插件 release ID
	NodeKey         any         // 节点唯一标识
	DesiredState    any         // 节点期望状态
	CurrentState    any         // 节点当前状态
	Generation      any         // 插件代际号
	LastHeartbeatAt *gtime.Time // 最近一次心跳时间
	ErrorMessage    any         // 节点错误信息
	CreatedAt       *gtime.Time // 创建时间
	UpdatedAt       *gtime.Time // 更新时间
}
