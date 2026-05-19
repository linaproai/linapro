// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysPluginState is the golang structure of table sys_plugin_state for DAO operations like Where/Data.
type SysPluginState struct {
	g.Meta     `orm:"table:sys_plugin_state, do:true"`
	Id         any        // Primary key ID
	PluginId   any        // Plugin unique identifier (kebab-case)
	TenantId   any        // Plugin state tenant ID, 0 means platform/global state
	StateKey   any        // State key
	StateValue any        // State value with JSON support
	Enabled    any        // Whether the plugin is enabled for the tenant
	CreatedAt  *time.Time // Creation time
	UpdatedAt  *time.Time // Update time
}
