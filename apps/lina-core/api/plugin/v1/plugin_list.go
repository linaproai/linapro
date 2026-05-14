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
	DependencyCheck         *PluginDependencyCheckResult `json:"dependencyCheck,omitempty" dc:"Server-side dependency check result used by management UI for install/uninstall planning" eg:"{}"`
	RequestedHostServices   []*HostServicePermissionItem `json:"requestedHostServices,omitempty" dc:"The hostServices application list declared by the current version of the plugin" eg:"[]"`
	AuthorizedHostServices  []*HostServicePermissionItem `json:"authorizedHostServices,omitempty" dc:"HostServices authorization snapshot after final confirmation of the current release by the host" eg:"[]"`
	DeclaredRoutes          []*PluginRouteReviewItem     `json:"declaredRoutes,omitempty" dc:"The dynamic route review list declared by the current release, only returned by dynamic plugins; used to display the real public routes that the plugin will expose before installation or activation." eg:"[]"`
}

// PluginDependencyCheckResult describes one server-side plugin dependency decision.
type PluginDependencyCheckResult struct {
	TargetId              string                              `json:"targetId" dc:"Checked plugin ID" eg:"plugin-demo-dynamic"`
	Framework             PluginDependencyFrameworkCheck      `json:"framework" dc:"Framework compatibility check result" eg:"{}"`
	Dependencies          []*PluginDependencyItem             `json:"dependencies" dc:"Direct and transitive plugin dependency checks" eg:"[]"`
	AutoInstallPlan       []*PluginDependencyAutoInstallItem  `json:"autoInstallPlan" dc:"Dependency plugins that will be installed automatically before the target" eg:"[]"`
	AutoInstalled         []*PluginDependencyAutoInstallItem  `json:"autoInstalled,omitempty" dc:"Dependency plugins installed automatically during the current install request" eg:"[]"`
	ManualInstallRequired []*PluginDependencyItem             `json:"manualInstallRequired" dc:"Hard dependencies that must be installed manually first" eg:"[]"`
	SoftUnsatisfied       []*PluginDependencyItem             `json:"softUnsatisfied" dc:"Soft dependencies that are currently missing or incompatible" eg:"[]"`
	Blockers              []*PluginDependencyBlocker          `json:"blockers" dc:"Install-side hard dependency blockers that prevent install or upgrade side effects" eg:"[]"`
	Cycle                 []string                            `json:"cycle,omitempty" dc:"Detected dependency cycle chain" eg:"[]"`
	ReverseDependents     []*PluginDependencyReverseDependent `json:"reverseDependents" dc:"Installed downstream plugins depending on this plugin" eg:"[]"`
	ReverseBlockers       []*PluginDependencyBlocker          `json:"reverseBlockers" dc:"Uninstall or downstream-version blockers for reverse dependency protection" eg:"[]"`
}

// PluginDependencyFrameworkCheck describes framework version compatibility.
type PluginDependencyFrameworkCheck struct {
	RequiredVersion string `json:"requiredVersion" dc:"Declared framework version range" eg:">=0.1.0 <1.0.0"`
	CurrentVersion  string `json:"currentVersion" dc:"Current LinaPro framework version" eg:"v0.1.0"`
	Status          string `json:"status" dc:"Compatibility state: not_declared, satisfied, unsatisfied" eg:"satisfied"`
}

// PluginDependencyItem describes one plugin dependency edge.
type PluginDependencyItem struct {
	OwnerId         string   `json:"ownerId" dc:"Plugin declaring the dependency" eg:"plugin-demo-dynamic"`
	DependencyId    string   `json:"dependencyId" dc:"Depended-on plugin ID" eg:"plugin-demo-source"`
	DependencyName  string   `json:"dependencyName" dc:"Depended-on plugin display name" eg:"Source Plugin Demo"`
	RequiredVersion string   `json:"requiredVersion" dc:"Declared dependency version range" eg:">=0.1.0"`
	CurrentVersion  string   `json:"currentVersion" dc:"Discovered or installed dependency version" eg:"v0.1.0"`
	Required        bool     `json:"required" dc:"Whether this dependency blocks lifecycle when unsatisfied" eg:"true"`
	InstallMode     string   `json:"installMode" dc:"Declared install strategy: manual or auto" eg:"auto"`
	Installed       bool     `json:"installed" dc:"Whether the dependency plugin is installed" eg:"false"`
	Discovered      bool     `json:"discovered" dc:"Whether the dependency plugin was found in the catalog" eg:"true"`
	Status          string   `json:"status" dc:"Dependency state from the server resolver" eg:"auto_install_planned"`
	Chain           []string `json:"chain,omitempty" dc:"Dependency chain leading to this edge" eg:"[]"`
}

// PluginDependencyAutoInstallItem describes one automatic dependency install.
type PluginDependencyAutoInstallItem struct {
	PluginId   string   `json:"pluginId" dc:"Dependency plugin ID" eg:"plugin-demo-source"`
	Name       string   `json:"name" dc:"Dependency plugin display name" eg:"Source Plugin Demo"`
	Version    string   `json:"version" dc:"Dependency version to install" eg:"v0.1.0"`
	RequiredBy string   `json:"requiredBy" dc:"Direct parent plugin requesting the dependency" eg:"plugin-demo-dynamic"`
	Chain      []string `json:"chain,omitempty" dc:"Dependency chain leading to this plan item" eg:"[]"`
}

// PluginDependencyBlocker describes one hard dependency failure.
type PluginDependencyBlocker struct {
	Code            string   `json:"code" dc:"Blocker category" eg:"dependency_missing"`
	PluginId        string   `json:"pluginId" dc:"Plugin whose lifecycle is blocked" eg:"plugin-demo-dynamic"`
	DependencyId    string   `json:"dependencyId" dc:"Dependency plugin when applicable" eg:"plugin-demo-source"`
	RequiredVersion string   `json:"requiredVersion" dc:"Declared version range when applicable" eg:">=0.1.0"`
	CurrentVersion  string   `json:"currentVersion" dc:"Observed version when applicable" eg:"v0.1.0"`
	Chain           []string `json:"chain,omitempty" dc:"Dependency chain associated with this blocker" eg:"[]"`
	Detail          string   `json:"detail" dc:"Concise operator diagnostic" eg:"dependency_missing"`
}

// PluginDependencyReverseDependent describes one installed downstream dependency.
type PluginDependencyReverseDependent struct {
	PluginId        string `json:"pluginId" dc:"Downstream plugin ID" eg:"content-notice"`
	Name            string `json:"name" dc:"Downstream plugin display name" eg:"Content Notice"`
	Version         string `json:"version" dc:"Downstream plugin version" eg:"v0.1.0"`
	RequiredVersion string `json:"requiredVersion" dc:"Version range declared by downstream plugin" eg:">=0.1.0"`
}
