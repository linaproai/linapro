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
	Id         any         // Primary key ID
	PluginId   any         // Plugin unique identifier (kebab-case)
	StateKey   any         // State key
	StateValue any         // State value with JSON support
	CreatedAt  *gtime.Time // Creation time
	UpdatedAt  *gtime.Time // Update time
}
