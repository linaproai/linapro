package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq is the request for querying plugin list.
type ListReq struct {
	g.Meta    `path:"/plugins" method:"get" tags:"Plugin Management" summary:"Query plugin list" permission:"plugin:query" dc:"Scan the source plugin directory and synchronize the basic status of the plugin, and return the plugin list and activation status"`
	Id        string `json:"id" dc:"Filter by the unique identifier of the plugin, fuzzy matching, query all if not passed" eg:"plugin-demo-source"`
	Name      string `json:"name" dc:"Filter by plugin name, fuzzy match, query all if not passed" eg:"Source Plugin Demo"`
	Type      string `json:"type" dc:"Filter by plugin type: source=source plugin dynamic=dynamic plugin, if not passed, all will be queried; the current dynamic plugin implementation only supports WASM" eg:"dynamic"`
	Status    *int   `json:"status" dc:"Filter by enabled status: 1=enabled 0=disabled, if not passed, query all" eg:"1"`
	Installed *int   `json:"installed" dc:"Filter by installation status: 1=Installed 0=Not installed, if not uploaded, query all" eg:"1"`
}

// ListRes is the response for querying plugin list.
type ListRes struct {
	List  []*PluginItem `json:"list" dc:"Plugin list" eg:"[]"`
	Total int           `json:"total" dc:"Total number of plugins" eg:"1"`
}

// PluginItem represents plugin information.
type PluginItem struct {
	Id                      string                       `json:"id" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
	Name                    string                       `json:"name" dc:"Plugin name" eg:"Source Plugin Demo"`
	Version                 string                       `json:"version" dc:"Plugin current manifest version number" eg:"v0.1.0"`
	Type                    string                       `json:"type" dc:"Plugin first-level type: source=source plugin dynamic=dynamic plugin" eg:"source"`
	Description             string                       `json:"description" dc:"Plugin description" eg:"Source plugin that provides left-side menu pages and public/protected routing examples"`
	Installed               int                          `json:"installed" dc:"Installation status: 1=installed 0=not installed; the source plugin can still be in the uninstalled state by default after being discovered by the host" eg:"1"`
	InstalledAt             string                       `json:"installedAt" dc:"Plugin installation time, returns an empty string if it is not installed." eg:"2026-01-01 12:00:00"`
	Enabled                 int                          `json:"enabled" dc:"Enabled status: 1=enabled 0=disabled" eg:"1"`
	AutoEnableManaged       int                          `json:"autoEnableManaged" dc:"Whether it is hit by plugin.autoEnable in the host's main configuration file: 1=yes 0=no; if hit, it means that the host will ensure that the plugin is enabled when it starts." eg:"1"`
	AutoEnableForNewTenants bool                         `json:"autoEnableForNewTenants" dc:"Platform policy: whether installed and enabled tenant-scoped plugins are enabled for new tenants automatically" eg:"true"`
	SupportsMultiTenant     bool                         `json:"supportsMultiTenant" dc:"Whether the plugin manifest declares support for tenant-level plugin governance" eg:"true"`
	ScopeNature             string                       `json:"scopeNature" dc:"Plugin scope nature: platform_only or tenant_aware" eg:"tenant_aware"`
	InstallMode             string                       `json:"installMode" dc:"Plugin install mode: global or tenant_scoped" eg:"tenant_scoped"`
	StatusKey               string                       `json:"statusKey" dc:"The location key name of the plugin status in the system plugin registry. The frontend registry monitor will use this key to determine whether the plugin status needs to be refreshed." eg:"sys_plugin.status:plugin-demo-source"`
	UpdatedAt               string                       `json:"updatedAt" dc:"Plugin registry last updated time" eg:"2026-01-01 12:00:00"`
	AuthorizationRequired   int                          `json:"authorizationRequired" dc:"Whether there is a hostServices resource application that needs to be confirmed during installation/activation: 1=Yes 0=No" eg:"1"`
	AuthorizationStatus     string                       `json:"authorizationStatus" dc:"Current authorization status: not_required=no confirmation required pending=to be confirmed confirmed=confirmed" eg:"confirmed"`
	HasMockData             int                          `json:"hasMockData" dc:"Whether the plugin ships any mock-data SQL files under manifest/sql/mock-data/: 1=yes 0=no. The frontend uses this to decide whether to render the optional Install mock data checkbox in the install dialog." eg:"1"`
	RequestedHostServices   []*HostServicePermissionItem `json:"requestedHostServices,omitempty" dc:"The hostServices application list declared by the current version of the plugin" eg:"[]"`
	AuthorizedHostServices  []*HostServicePermissionItem `json:"authorizedHostServices,omitempty" dc:"HostServices authorization snapshot after final confirmation of the current release by the host" eg:"[]"`
	DeclaredRoutes          []*PluginRouteReviewItem     `json:"declaredRoutes,omitempty" dc:"The dynamic route review list declared by the current release, only returned by dynamic plugins; used to display the real public routes that the plugin will expose before installation or activation." eg:"[]"`
}
