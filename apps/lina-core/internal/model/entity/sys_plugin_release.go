// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPluginRelease is the golang structure for table sys_plugin_release.
type SysPluginRelease struct {
	Id               int         `json:"id"               orm:"id"                description:"主键ID"`
	PluginId         string      `json:"pluginId"         orm:"plugin_id"         description:"插件唯一标识（kebab-case）"`
	ReleaseVersion   string      `json:"releaseVersion"   orm:"release_version"   description:"插件版本号"`
	Type             string      `json:"type"             orm:"type"              description:"插件一级类型（source/dynamic）"`
	RuntimeKind      string      `json:"runtimeKind"      orm:"runtime_kind"      description:"运行时产物类型（当前仅 wasm）"`
	SchemaVersion    string      `json:"schemaVersion"    orm:"schema_version"    description:"plugin.yaml 清单 schema 版本"`
	MinHostVersion   string      `json:"minHostVersion"   orm:"min_host_version"  description:"宿主最小兼容版本"`
	MaxHostVersion   string      `json:"maxHostVersion"   orm:"max_host_version"  description:"宿主最大兼容版本"`
	Status           string      `json:"status"           orm:"status"            description:"release 状态（prepared/installed/active/uninstalled/failed）"`
	ManifestPath     string      `json:"manifestPath"     orm:"manifest_path"     description:"插件清单路径"`
	PackagePath      string      `json:"packagePath"      orm:"package_path"      description:"插件源码目录或运行时产物路径"`
	Checksum         string      `json:"checksum"         orm:"checksum"          description:"插件清单或产物校验值"`
	ManifestSnapshot string      `json:"manifestSnapshot" orm:"manifest_snapshot" description:"插件清单与资源摘要快照（YAML，不保存具体 SQL/前端文件路径）"`
	CreatedAt        *gtime.Time `json:"createdAt"        orm:"created_at"        description:"创建时间"`
	UpdatedAt        *gtime.Time `json:"updatedAt"        orm:"updated_at"        description:"更新时间"`
	DeletedAt        *gtime.Time `json:"deletedAt"        orm:"deleted_at"        description:"删除时间"`
}
