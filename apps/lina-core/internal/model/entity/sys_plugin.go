// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPlugin is the golang structure for table sys_plugin.
type SysPlugin struct {
	Id           int         `json:"id"           orm:"id"            description:"主键ID"`
	PluginId     string      `json:"pluginId"     orm:"plugin_id"     description:"插件唯一标识（kebab-case）"`
	Name         string      `json:"name"         orm:"name"          description:"插件名称"`
	Version      string      `json:"version"      orm:"version"       description:"插件版本号"`
	Type         string      `json:"type"         orm:"type"          description:"插件一级类型（source/dynamic）"`
	Installed    int         `json:"installed"    orm:"installed"     description:"安装状态（1=已安装 0=未安装）"`
	Status       int         `json:"status"       orm:"status"        description:"启用状态（1=启用 0=禁用）"`
	DesiredState string      `json:"desiredState" orm:"desired_state" description:"宿主期望状态（uninstalled/installed/enabled）"`
	CurrentState string      `json:"currentState" orm:"current_state" description:"宿主当前状态（uninstalled/installed/enabled/reconciling/failed）"`
	Generation   int64       `json:"generation"   orm:"generation"    description:"宿主当前生效代际号"`
	ReleaseId    int         `json:"releaseId"    orm:"release_id"    description:"宿主当前生效 release ID"`
	ManifestPath string      `json:"manifestPath" orm:"manifest_path" description:"插件清单文件路径"`
	Checksum     string      `json:"checksum"     orm:"checksum"      description:"插件包校验值"`
	InstalledAt  *gtime.Time `json:"installedAt"  orm:"installed_at"  description:"安装时间"`
	EnabledAt    *gtime.Time `json:"enabledAt"    orm:"enabled_at"    description:"最后一次启用时间"`
	DisabledAt   *gtime.Time `json:"disabledAt"   orm:"disabled_at"   description:"最后一次禁用时间"`
	Remark       string      `json:"remark"       orm:"remark"        description:"备注"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:"创建时间"`
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:"更新时间"`
	DeletedAt    *gtime.Time `json:"deletedAt"    orm:"deleted_at"    description:"删除时间"`
}
