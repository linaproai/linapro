// This file defines the plugin runtime-upgrade preview API DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpgradePreviewReq is the request for previewing one plugin runtime upgrade.
type UpgradePreviewReq struct {
	g.Meta `path:"/plugins/{id}/upgrade/preview" method:"get" tags:"Plugin Management" summary:"Preview plugin runtime upgrade" permission:"plugin:query" dc:"Return a side-effect-free runtime upgrade preview for a pending-upgrade or failed-upgrade plugin, including effective and target manifest snapshots, dependency checks, SQL summary, hostServices changes, and risk hints. This endpoint does not execute SQL, switch releases, update state, or invoke plugin callbacks."`
	Id     string `json:"id" v:"required|length:1,64" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
}

// UpgradePreviewRes is the response for previewing one plugin runtime upgrade.
type UpgradePreviewRes struct {
	PluginId          string                        `json:"pluginId" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
	RuntimeState      string                        `json:"runtimeState" dc:"Current plugin runtime upgrade state, expected to be pending_upgrade or upgrade_failed for preview" eg:"pending_upgrade"`
	EffectiveVersion  string                        `json:"effectiveVersion" dc:"Database-effective plugin version before upgrade" eg:"v0.1.0"`
	DiscoveredVersion string                        `json:"discoveredVersion" dc:"Target plugin version discovered from source plugin.yaml or dynamic artifact metadata" eg:"v0.2.0"`
	FromManifest      *PluginManifestSnapshotItem   `json:"fromManifest" dc:"Current effective manifest snapshot before upgrade" eg:"{}"`
	ToManifest        *PluginManifestSnapshotItem   `json:"toManifest" dc:"Target manifest snapshot after upgrade" eg:"{}"`
	DependencyCheck   *PluginDependencyCheckResult  `json:"dependencyCheck" dc:"Server-side dependency check result for the target upgrade" eg:"{}"`
	SQLSummary        PluginUpgradeSQLSummary       `json:"sqlSummary" dc:"SQL asset count summary for the target manifest; preview does not execute these files" eg:"{}"`
	HostServicesDiff  PluginUpgradeHostServicesDiff `json:"hostServicesDiff" dc:"Service-level hostServices change summary between effective and target manifests" eg:"{}"`
	RiskHints         []string                      `json:"riskHints" dc:"Stable i18n keys for operator-facing upgrade risk hints" eg:"[\"plugin.runtimeUpgrade.risk.upgradeSqlRequiresReview\"]"`
}

// PluginManifestSnapshotItem stores review-friendly manifest metadata for upgrade preview.
type PluginManifestSnapshotItem struct {
	Id                        string                       `json:"id" dc:"Plugin unique identifier" eg:"plugin-demo-source"`
	Name                      string                       `json:"name" dc:"Plugin display name" eg:"Source Plugin Demo"`
	Version                   string                       `json:"version" dc:"Manifest version" eg:"v0.2.0"`
	Type                      string                       `json:"type" dc:"Plugin type: source or dynamic" eg:"source"`
	ScopeNature               string                       `json:"scopeNature,omitempty" dc:"Plugin scope nature: platform_only or tenant_aware" eg:"tenant_aware"`
	SupportsMultiTenant       bool                         `json:"supportsMultiTenant" dc:"Whether the manifest supports tenant-level plugin governance" eg:"true"`
	DefaultInstallMode        string                       `json:"defaultInstallMode,omitempty" dc:"Default plugin install mode: global or tenant_scoped" eg:"tenant_scoped"`
	Description               string                       `json:"description,omitempty" dc:"Plugin description" eg:"Source plugin that provides examples"`
	RuntimeKind               string                       `json:"runtimeKind,omitempty" dc:"Dynamic runtime kind when present" eg:"wasm"`
	RuntimeAbiVersion         string                       `json:"runtimeAbiVersion,omitempty" dc:"Dynamic runtime ABI version when present" eg:"v1"`
	ManifestDeclared          bool                         `json:"manifestDeclared" dc:"Whether the manifest is backed by plugin.yaml or embedded artifact metadata" eg:"true"`
	InstallSQLCount           int                          `json:"installSqlCount" dc:"Number of install/upgrade SQL assets visible in the manifest" eg:"1"`
	UninstallSQLCount         int                          `json:"uninstallSqlCount" dc:"Number of uninstall SQL assets visible in the manifest" eg:"1"`
	MockSQLCount              int                          `json:"mockSqlCount" dc:"Number of mock SQL assets visible in the manifest; runtime upgrade excludes mock data" eg:"0"`
	FrontendPageCount         int                          `json:"frontendPageCount" dc:"Number of source frontend page entries" eg:"1"`
	FrontendSlotCount         int                          `json:"frontendSlotCount" dc:"Number of source frontend slot entries" eg:"0"`
	MenuCount                 int                          `json:"menuCount" dc:"Number of manifest menu entries" eg:"2"`
	BackendHookCount          int                          `json:"backendHookCount" dc:"Number of backend hook declarations" eg:"1"`
	ResourceSpecCount         int                          `json:"resourceSpecCount" dc:"Number of backend resource declarations" eg:"1"`
	RouteCount                int                          `json:"routeCount" dc:"Number of dynamic route declarations" eg:"1"`
	RouteExecutionEnabled     bool                         `json:"routeExecutionEnabled" dc:"Whether dynamic route execution is enabled in the bridge metadata" eg:"true"`
	RouteRequestCodec         string                       `json:"routeRequestCodec,omitempty" dc:"Dynamic route request codec" eg:"protobuf"`
	RouteResponseCodec        string                       `json:"routeResponseCodec,omitempty" dc:"Dynamic route response codec" eg:"protobuf"`
	RuntimeFrontendAssetCount int                          `json:"runtimeFrontendAssetCount" dc:"Number of frontend assets embedded in a dynamic artifact" eg:"2"`
	RuntimeSQLAssetCount      int                          `json:"runtimeSqlAssetCount" dc:"Number of SQL assets embedded in a dynamic artifact" eg:"1"`
	HostServiceAuthRequired   bool                         `json:"hostServiceAuthRequired" dc:"Whether the manifest requests resource-scoped hostServices authorization" eg:"true"`
	HostServiceAuthConfirmed  bool                         `json:"hostServiceAuthConfirmed" dc:"Whether the snapshot already carries host-confirmed authorization" eg:"false"`
	RequestedHostServices     []*HostServicePermissionItem `json:"requestedHostServices,omitempty" dc:"HostServices requested by this manifest snapshot" eg:"[]"`
	AuthorizedHostServices    []*HostServicePermissionItem `json:"authorizedHostServices,omitempty" dc:"Host-confirmed hostServices stored on this manifest snapshot" eg:"[]"`
}

// PluginUpgradeSQLSummary summarizes target manifest SQL assets.
type PluginUpgradeSQLSummary struct {
	InstallSQLCount      int `json:"installSqlCount" dc:"Number of install/upgrade SQL assets in the target manifest" eg:"1"`
	UninstallSQLCount    int `json:"uninstallSqlCount" dc:"Number of uninstall SQL assets in the target manifest" eg:"1"`
	MockSQLCount         int `json:"mockSqlCount" dc:"Number of mock SQL assets in the target manifest; runtime upgrade excludes mock data" eg:"0"`
	RuntimeSQLAssetCount int `json:"runtimeSqlAssetCount" dc:"Number of SQL assets embedded in a dynamic artifact" eg:"1"`
}

// PluginUpgradeHostServicesDiff summarizes hostServices changes for upgrade preview.
type PluginUpgradeHostServicesDiff struct {
	Added                 []*PluginUpgradeHostServiceChange `json:"added" dc:"Host services added by the target manifest" eg:"[]"`
	Removed               []*PluginUpgradeHostServiceChange `json:"removed" dc:"Host services removed by the target manifest" eg:"[]"`
	Changed               []*PluginUpgradeHostServiceChange `json:"changed" dc:"Host services whose methods or governed targets changed" eg:"[]"`
	AuthorizationRequired bool                              `json:"authorizationRequired" dc:"Whether the target manifest requires hostServices authorization confirmation" eg:"true"`
	AuthorizationChanged  bool                              `json:"authorizationChanged" dc:"Whether requested hostServices differ between effective and target manifests" eg:"true"`
}

// PluginUpgradeHostServiceChange summarizes one changed host service.
type PluginUpgradeHostServiceChange struct {
	Service           string   `json:"service" dc:"Host service identifier" eg:"storage"`
	FromMethods       []string `json:"fromMethods,omitempty" dc:"Effective method set before upgrade" eg:"[\"get\"]"`
	ToMethods         []string `json:"toMethods,omitempty" dc:"Target method set after upgrade" eg:"[\"get\",\"put\"]"`
	FromResourceCount int      `json:"fromResourceCount" dc:"Effective governed resource-ref count before upgrade" eg:"0"`
	ToResourceCount   int      `json:"toResourceCount" dc:"Target governed resource-ref count after upgrade" eg:"1"`
	FromTables        []string `json:"fromTables,omitempty" dc:"Effective data-table set before upgrade" eg:"[]"`
	ToTables          []string `json:"toTables,omitempty" dc:"Target data-table set after upgrade" eg:"[]"`
	FromPaths         []string `json:"fromPaths,omitempty" dc:"Effective storage path set before upgrade" eg:"[\"reports/\"]"`
	ToPaths           []string `json:"toPaths,omitempty" dc:"Target storage path set after upgrade" eg:"[\"reports/\",\"exports/\"]"`
}
