// This file declares tenant plugin-governance list DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// TenantPluginEntity is the tenant-facing plugin governance projection.
type TenantPluginEntity struct {
	Id            string `json:"id" dc:"Plugin unique identifier" eg:"monitor-loginlog"`
	Name          string `json:"name" dc:"Plugin display name" eg:"Login Log"`
	Version       string `json:"version" dc:"Plugin version" eg:"v0.1.0"`
	Type          string `json:"type" dc:"Plugin type: source or dynamic" eg:"source"`
	Description   string `json:"description" dc:"Plugin description" eg:"Tenant login audit"`
	Installed     int    `json:"installed" dc:"Whether the plugin is installed: 1=yes 0=no" eg:"1"`
	Enabled       int    `json:"enabled" dc:"Whether the plugin is globally enabled: 1=yes 0=no" eg:"1"`
	ScopeNature   string `json:"scopeNature" dc:"Plugin scope nature: tenant_aware or platform_only" eg:"tenant_aware"`
	InstallMode   string `json:"installMode" dc:"Plugin install mode: global or tenant_scoped" eg:"tenant_scoped"`
	TenantEnabled int    `json:"tenantEnabled" dc:"Whether the plugin is enabled for the current tenant: 1=yes 0=no" eg:"1"`
}

// TenantPluginListReq defines the request for listing tenant-controllable plugins.
type TenantPluginListReq struct {
	g.Meta `path:"/tenant/plugins" method:"get" tags:"Tenant Plugins" summary:"Get tenant plugin list" dc:"List tenant-scoped plugins that the current tenant administrator may enable or disable." permission:"system:tenant:plugin:list"`
}

// TenantPluginListRes defines the tenant plugin list response.
type TenantPluginListRes struct {
	List  []*TenantPluginEntity `json:"list" dc:"Tenant plugin list" eg:"[]"`
	Total int                   `json:"total" dc:"Total number of tenant plugins" eg:"1"`
}
