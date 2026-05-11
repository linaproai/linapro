// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SysPlugin is the golang structure for table sys_plugin.
type SysPlugin struct {
	Id                      int         `json:"id"                      orm:"id"                          description:"Primary key ID"`
	PluginId                string      `json:"pluginId"                orm:"plugin_id"                   description:"Plugin unique identifier (kebab-case)"`
	Name                    string      `json:"name"                    orm:"name"                        description:"Plugin name"`
	Version                 string      `json:"version"                 orm:"version"                     description:"Plugin version"`
	Type                    string      `json:"type"                    orm:"type"                        description:"Plugin top-level type: source/dynamic"`
	ScopeNature             string      `json:"scopeNature"             orm:"scope_nature"                description:"Plugin scope nature: platform_only or tenant_aware"`
	InstallMode             string      `json:"installMode"             orm:"install_mode"                description:"Plugin install mode: global or tenant_scoped"`
	AutoEnableForNewTenants bool        `json:"autoEnableForNewTenants" orm:"auto_enable_for_new_tenants" description:"Platform policy: whether installed and enabled tenant-scoped plugins are enabled for new tenants automatically"`
	Installed               int         `json:"installed"               orm:"installed"                   description:"Installation status: 1=installed, 0=not installed"`
	Status                  int         `json:"status"                  orm:"status"                      description:"Enablement status: 1=enabled, 0=disabled"`
	DesiredState            string      `json:"desiredState"            orm:"desired_state"               description:"Host desired state: uninstalled/installed/enabled"`
	CurrentState            string      `json:"currentState"            orm:"current_state"               description:"Host current state: uninstalled/installed/enabled/reconciling/failed"`
	Generation              int64       `json:"generation"              orm:"generation"                  description:"Current host generation number"`
	ReleaseId               int         `json:"releaseId"               orm:"release_id"                  description:"Current active host release ID"`
	ManifestPath            string      `json:"manifestPath"            orm:"manifest_path"               description:"Plugin manifest file path"`
	Checksum                string      `json:"checksum"                orm:"checksum"                    description:"Plugin package checksum"`
	InstalledAt             *gtime.Time `json:"installedAt"             orm:"installed_at"                description:"Installation time"`
	EnabledAt               *gtime.Time `json:"enabledAt"               orm:"enabled_at"                  description:"Last enabled time"`
	DisabledAt              *gtime.Time `json:"disabledAt"              orm:"disabled_at"                 description:"Last disabled time"`
	Remark                  string      `json:"remark"                  orm:"remark"                      description:"Remark"`
	CreatedAt               *gtime.Time `json:"createdAt"               orm:"created_at"                  description:"Creation time"`
	UpdatedAt               *gtime.Time `json:"updatedAt"               orm:"updated_at"                  description:"Update time"`
	DeletedAt               *gtime.Time `json:"deletedAt"               orm:"deleted_at"                  description:"Deletion time"`
}
