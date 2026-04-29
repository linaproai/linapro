// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPluginResourceRef is the golang structure for table sys_plugin_resource_ref.
type SysPluginResourceRef struct {
	Id           int         `json:"id"           orm:"id"            description:"Primary key ID"`
	PluginId     string      `json:"pluginId"     orm:"plugin_id"     description:"Plugin unique identifier (kebab-case)"`
	ReleaseId    int         `json:"releaseId"    orm:"release_id"    description:"Owning plugin release ID"`
	ResourceType string      `json:"resourceType" orm:"resource_type" description:"Resource type: manifest/sql/frontend/menu/permission, etc."`
	ResourceKey  string      `json:"resourceKey"  orm:"resource_key"  description:"Resource unique key"`
	ResourcePath string      `json:"resourcePath" orm:"resource_path" description:"Resource location metadata, empty by default and without concrete frontend or SQL paths"`
	OwnerType    string      `json:"ownerType"    orm:"owner_type"    description:"Host object type: file/menu/route/slot, etc."`
	OwnerKey     string      `json:"ownerKey"     orm:"owner_key"     description:"Stable host object identifier"`
	Remark       string      `json:"remark"       orm:"remark"        description:"Remark"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:"Creation time"`
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:"Update time"`
	DeletedAt    *gtime.Time `json:"deletedAt"    orm:"deleted_at"    description:"Deletion time"`
}
