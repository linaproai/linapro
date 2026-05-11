// This file declares tenant plugin disable DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// TenantPluginDisableReq defines the request for disabling one tenant-scoped plugin.
type TenantPluginDisableReq struct {
	g.Meta   `path:"/tenant/plugins/{pluginId}/disable" method:"post" tags:"Tenant Plugins" summary:"Disable tenant plugin" dc:"Disable one tenant-scoped plugin for the current tenant." permission:"system:tenant:plugin:disable"`
	PluginId string `json:"pluginId" v:"required|length:1,64" dc:"Plugin unique identifier" eg:"monitor-loginlog"`
}

// TenantPluginDisableRes defines the tenant plugin disable response.
type TenantPluginDisableRes struct{}
