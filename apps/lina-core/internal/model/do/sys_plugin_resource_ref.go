// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysPluginResourceRef is the golang structure of table sys_plugin_resource_ref for DAO operations like Where/Data.
type SysPluginResourceRef struct {
	g.Meta       `orm:"table:sys_plugin_resource_ref, do:true"`
	Id           any        // Primary key ID
	PluginId     any        // Plugin unique identifier (kebab-case)
	ReleaseId    any        // Owning plugin release ID
	ResourceType any        // Resource type: manifest/sql/frontend/menu/permission, etc.
	ResourceKey  any        // Resource unique key
	ResourcePath any        // Resource location metadata, empty by default and without concrete frontend or SQL paths
	OwnerType    any        // Host object type: file/menu/route/slot, etc.
	OwnerKey     any        // Stable host object identifier
	Remark       any        // Remark
	CreatedAt    *time.Time // Creation time
	UpdatedAt    *time.Time // Update time
	DeletedAt    *time.Time // Deletion time
}
