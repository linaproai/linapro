// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPluginState is the golang structure for table sys_plugin_state.
type SysPluginState struct {
	Id         int         `json:"id"         orm:"id"          description:"Primary key ID"`
	PluginId   string      `json:"pluginId"   orm:"plugin_id"   description:"Plugin unique identifier (kebab-case)"`
	TenantId   int         `json:"tenantId"   orm:"tenant_id"   description:"Plugin state tenant ID, 0 means platform/global state"`
	StateKey   string      `json:"stateKey"   orm:"state_key"   description:"State key"`
	StateValue string      `json:"stateValue" orm:"state_value" description:"State value with JSON support"`
	Enabled    bool        `json:"enabled"    orm:"enabled"     description:"Whether the plugin is enabled for the tenant"`
	CreatedAt  *gtime.Time `json:"createdAt"  orm:"created_at"  description:"Creation time"`
	UpdatedAt  *gtime.Time `json:"updatedAt"  orm:"updated_at"  description:"Update time"`
}
