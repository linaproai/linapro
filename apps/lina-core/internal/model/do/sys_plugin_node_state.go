// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysPluginNodeState is the golang structure of table sys_plugin_node_state for DAO operations like Where/Data.
type SysPluginNodeState struct {
	g.Meta          `orm:"table:sys_plugin_node_state, do:true"`
	Id              any        // Primary key ID
	PluginId        any        // Plugin unique identifier (kebab-case)
	ReleaseId       any        // Owning plugin release ID
	NodeKey         any        // Node unique identifier
	DesiredState    any        // Node desired state
	CurrentState    any        // Node current state
	Generation      any        // Plugin generation number
	LastHeartbeatAt *time.Time // Last heartbeat time
	ErrorMessage    any        // Node error message
	CreatedAt       *time.Time // Creation time
	UpdatedAt       *time.Time // Update time
}
