// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPluginNodeState is the golang structure for table sys_plugin_node_state.
type SysPluginNodeState struct {
	Id              int         `json:"id"              orm:"id"                description:"Primary key ID"`
	PluginId        string      `json:"pluginId"        orm:"plugin_id"         description:"Plugin unique identifier (kebab-case)"`
	ReleaseId       int         `json:"releaseId"       orm:"release_id"        description:"Owning plugin release ID"`
	NodeKey         string      `json:"nodeKey"         orm:"node_key"          description:"Node unique identifier"`
	DesiredState    string      `json:"desiredState"    orm:"desired_state"     description:"Node desired state"`
	CurrentState    string      `json:"currentState"    orm:"current_state"     description:"Node current state"`
	Generation      int64       `json:"generation"      orm:"generation"        description:"Plugin generation number"`
	LastHeartbeatAt *gtime.Time `json:"lastHeartbeatAt" orm:"last_heartbeat_at" description:"Last heartbeat time"`
	ErrorMessage    string      `json:"errorMessage"    orm:"error_message"     description:"Node error message"`
	CreatedAt       *gtime.Time `json:"createdAt"       orm:"created_at"        description:"Creation time"`
	UpdatedAt       *gtime.Time `json:"updatedAt"       orm:"updated_at"        description:"Update time"`
}
