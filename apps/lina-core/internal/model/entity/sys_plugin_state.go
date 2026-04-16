// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPluginState is the golang structure for table sys_plugin_state.
type SysPluginState struct {
	Id         int         `json:"id"         orm:"id"          description:"主键ID"`
	PluginId   string      `json:"pluginId"   orm:"plugin_id"   description:"插件唯一标识（kebab-case）"`
	StateKey   string      `json:"stateKey"   orm:"state_key"   description:"状态键"`
	StateValue string      `json:"stateValue" orm:"state_value" description:"状态值（支持JSON）"`
	CreatedAt  *gtime.Time `json:"createdAt"  orm:"created_at"  description:"创建时间"`
	UpdatedAt  *gtime.Time `json:"updatedAt"  orm:"updated_at"  description:"更新时间"`
}
