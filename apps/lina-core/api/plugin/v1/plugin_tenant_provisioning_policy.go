package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateTenantProvisioningPolicyReq updates the platform-owned tenant provisioning policy.
type UpdateTenantProvisioningPolicyReq struct {
	g.Meta                  `path:"/plugins/{id}/tenant-provisioning-policy" method:"put" tags:"Plugin Management" summary:"Update plugin tenant provisioning policy" permission:"plugin:edit" dc:"Update whether an installed and enabled tenant-scoped plugin is automatically enabled for newly created tenants. This is a platform policy and is not read from plugin.yaml."`
	Id                      string `json:"id" v:"required|length:1,64" dc:"Plugin unique identifier" eg:"content-notice"`
	AutoEnableForNewTenants bool   `json:"autoEnableForNewTenants" dc:"Whether newly created tenants should automatically receive this tenant-scoped plugin when the plugin is installed and enabled" eg:"true"`
}

// UpdateTenantProvisioningPolicyRes is the response for updating the tenant provisioning policy.
type UpdateTenantProvisioningPolicyRes struct {
	Id                      string `json:"id" dc:"Plugin unique identifier" eg:"content-notice"`
	AutoEnableForNewTenants bool   `json:"autoEnableForNewTenants" dc:"Whether newly created tenants should automatically receive this tenant-scoped plugin when the plugin is installed and enabled" eg:"true"`
}
