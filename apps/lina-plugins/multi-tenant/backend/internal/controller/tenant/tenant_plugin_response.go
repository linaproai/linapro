// This file implements tenant-scoped plugin response conversion helpers.

package tenant

import (
	v1 "lina-plugin-multi-tenant/backend/api/tenant/v1"
	"lina-plugin-multi-tenant/backend/internal/service/tenantplugin"
)

// toAPITenantPlugin converts a service plugin projection into an API DTO.
func toAPITenantPlugin(item *tenantplugin.Entity) *v1.TenantPluginEntity {
	if item == nil {
		return nil
	}
	return &v1.TenantPluginEntity{
		Id:            item.Id,
		Name:          item.Name,
		Version:       item.Version,
		Type:          item.Type,
		Description:   item.Description,
		Installed:     item.Installed,
		Enabled:       item.Enabled,
		ScopeNature:   item.ScopeNature,
		InstallMode:   item.InstallMode,
		TenantEnabled: item.TenantEnabled,
	}
}
