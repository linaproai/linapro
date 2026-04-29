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
	Id              any         // Primary key ID
	PluginId        any         // Plugin unique identifier (kebab-case)
	ReleaseId       any         // Owning plugin release ID
	NodeKey         any         // Node unique identifier
	DesiredState    any         // Node desired state
	CurrentState    any         // Node current state
	Generation      any         // Plugin generation number
	LastHeartbeatAt *gtime.Time // Last heartbeat time
	ErrorMessage    any         // Node error message
	CreatedAt       *gtime.Time // Creation time
	UpdatedAt       *gtime.Time // Update time
}
