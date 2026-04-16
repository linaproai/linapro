// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPluginResourceRef is the golang structure for table sys_plugin_resource_ref.
type SysPluginResourceRef struct {
	Id           int         `json:"id"           orm:"id"            description:"主键ID"`
	PluginId     string      `json:"pluginId"     orm:"plugin_id"     description:"插件唯一标识（kebab-case）"`
	ReleaseId    int         `json:"releaseId"    orm:"release_id"    description:"所属插件 release ID"`
	ResourceType string      `json:"resourceType" orm:"resource_type" description:"资源类型（manifest/sql/frontend/menu/permission 等）"`
	ResourceKey  string      `json:"resourceKey"  orm:"resource_key"  description:"资源唯一键"`
	ResourcePath string      `json:"resourcePath" orm:"resource_path" description:"资源定位补充信息（默认留空，不保存具体前端/SQL 路径）"`
	OwnerType    string      `json:"ownerType"    orm:"owner_type"    description:"宿主对象类型（file/menu/route/slot 等）"`
	OwnerKey     string      `json:"ownerKey"     orm:"owner_key"     description:"宿主对象稳定标识"`
	Remark       string      `json:"remark"       orm:"remark"        description:"备注"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:"创建时间"`
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:"更新时间"`
	DeletedAt    *gtime.Time `json:"deletedAt"    orm:"deleted_at"    description:"删除时间"`
}
