// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPlugin is the golang structure of table sys_plugin for DAO operations like Where/Data.
type SysPlugin struct {
	g.Meta                  `orm:"table:sys_plugin, do:true"`
	Id                      any         // Primary key ID
	PluginId                any         // Plugin unique identifier (kebab-case)
	Name                    any         // Plugin name
	Version                 any         // Plugin version
	Type                    any         // Plugin top-level type: source/dynamic
	ScopeNature             any         // Plugin scope nature: platform_only or tenant_aware
	InstallMode             any         // Plugin install mode: global or tenant_scoped
	AutoEnableForNewTenants any         // Platform policy: whether installed and enabled tenant-scoped plugins are enabled for new tenants automatically
	Installed               any         // Installation status: 1=installed, 0=not installed
	Status                  any         // Enablement status: 1=enabled, 0=disabled
	DesiredState            any         // Host desired state: uninstalled/installed/enabled
	CurrentState            any         // Host current state: uninstalled/installed/enabled/reconciling/failed
	Generation              any         // Current host generation number
	ReleaseId               any         // Current active host release ID
	ManifestPath            any         // Plugin manifest file path
	Checksum                any         // Plugin package checksum
	InstalledAt             *gtime.Time // Installation time
	EnabledAt               *gtime.Time // Last enabled time
	DisabledAt              *gtime.Time // Last disabled time
	Remark                  any         // Remark
	CreatedAt               *gtime.Time // Creation time
	UpdatedAt               *gtime.Time // Update time
	DeletedAt               *gtime.Time // Deletion time
}
