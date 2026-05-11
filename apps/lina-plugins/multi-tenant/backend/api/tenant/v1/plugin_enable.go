// This file declares tenant plugin enable DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// TenantPluginEnableReq defines the request for enabling one tenant-scoped plugin.
type TenantPluginEnableReq struct {
	g.Meta   `path:"/tenant/plugins/{pluginId}/enable" method:"post" tags:"Tenant Plugins" summary:"Enable tenant plugin" dc:"Enable one tenant-scoped plugin for the current tenant." permission:"system:tenant:plugin:enable"`
	PluginId string `json:"pluginId" v:"required|length:1,64" dc:"Plugin unique identifier" eg:"monitor-loginlog"`
}

// TenantPluginEnableRes defines the tenant plugin enable response.
type TenantPluginEnableRes struct{}
