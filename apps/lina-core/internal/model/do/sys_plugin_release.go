// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// SysPluginRelease is the golang structure of table sys_plugin_release for DAO operations like Where/Data.
type SysPluginRelease struct {
	g.Meta           `orm:"table:sys_plugin_release, do:true"`
	Id               any        // Primary key ID
	PluginId         any        // Plugin unique identifier (kebab-case)
	ReleaseVersion   any        // Plugin version
	Type             any        // Plugin top-level type: source/dynamic
	RuntimeKind      any        // Runtime artifact type (currently only wasm)
	SchemaVersion    any        // plugin.yaml manifest schema version
	MinHostVersion   any        // Minimum compatible host version
	MaxHostVersion   any        // Maximum compatible host version
	Status           any        // Release status: prepared/installed/active/uninstalled/failed
	ManifestPath     any        // Plugin manifest path
	PackagePath      any        // Plugin source directory or runtime artifact path
	Checksum         any        // Plugin manifest or artifact checksum
	ManifestSnapshot any        // Plugin manifest and resource summary snapshot in YAML, without concrete SQL or frontend file paths
	CreatedAt        *time.Time // Creation time
	UpdatedAt        *time.Time // Update time
	DeletedAt        *time.Time // Deletion time
}
