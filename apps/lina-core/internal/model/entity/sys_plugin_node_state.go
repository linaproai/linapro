// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPluginNodeState is the golang structure for table sys_plugin_node_state.
type SysPluginNodeState struct {
	Id              int         `json:"id"              orm:"id"                description:"主键ID"`
	PluginId        string      `json:"pluginId"        orm:"plugin_id"         description:"插件唯一标识（kebab-case）"`
	ReleaseId       int         `json:"releaseId"       orm:"release_id"        description:"所属插件 release ID"`
	NodeKey         string      `json:"nodeKey"         orm:"node_key"          description:"节点唯一标识"`
	DesiredState    string      `json:"desiredState"    orm:"desired_state"     description:"节点期望状态"`
	CurrentState    string      `json:"currentState"    orm:"current_state"     description:"节点当前状态"`
	Generation      int64       `json:"generation"      orm:"generation"        description:"插件代际号"`
	LastHeartbeatAt *gtime.Time `json:"lastHeartbeatAt" orm:"last_heartbeat_at" description:"最近一次心跳时间"`
	ErrorMessage    string      `json:"errorMessage"    orm:"error_message"     description:"节点错误信息"`
	CreatedAt       *gtime.Time `json:"createdAt"       orm:"created_at"        description:"创建时间"`
	UpdatedAt       *gtime.Time `json:"updatedAt"       orm:"updated_at"        description:"更新时间"`
}
