// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPluginRelease is the golang structure for table sys_plugin_release.
type SysPluginRelease struct {
	Id               int         `json:"id"               orm:"id"                description:"Primary key ID"`
	PluginId         string      `json:"pluginId"         orm:"plugin_id"         description:"Plugin unique identifier (kebab-case)"`
	ReleaseVersion   string      `json:"releaseVersion"   orm:"release_version"   description:"Plugin version"`
	Type             string      `json:"type"             orm:"type"              description:"Plugin top-level type: source/dynamic"`
	RuntimeKind      string      `json:"runtimeKind"      orm:"runtime_kind"      description:"Runtime artifact type (currently only wasm)"`
	SchemaVersion    string      `json:"schemaVersion"    orm:"schema_version"    description:"plugin.yaml manifest schema version"`
	MinHostVersion   string      `json:"minHostVersion"   orm:"min_host_version"  description:"Minimum compatible host version"`
	MaxHostVersion   string      `json:"maxHostVersion"   orm:"max_host_version"  description:"Maximum compatible host version"`
	Status           string      `json:"status"           orm:"status"            description:"Release status: prepared/installed/active/uninstalled/failed"`
	ManifestPath     string      `json:"manifestPath"     orm:"manifest_path"     description:"Plugin manifest path"`
	PackagePath      string      `json:"packagePath"      orm:"package_path"      description:"Plugin source directory or runtime artifact path"`
	Checksum         string      `json:"checksum"         orm:"checksum"          description:"Plugin manifest or artifact checksum"`
	ManifestSnapshot string      `json:"manifestSnapshot" orm:"manifest_snapshot" description:"Plugin manifest and resource summary snapshot in YAML, without concrete SQL or frontend file paths"`
	CreatedAt        *gtime.Time `json:"createdAt"        orm:"created_at"        description:"Creation time"`
	UpdatedAt        *gtime.Time `json:"updatedAt"        orm:"updated_at"        description:"Update time"`
	DeletedAt        *gtime.Time `json:"deletedAt"        orm:"deleted_at"        description:"Deletion time"`
}
