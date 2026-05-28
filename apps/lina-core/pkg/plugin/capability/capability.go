// Package capability defines the unified host capability services exposed to
// source plugins directly and to dynamic plugins through bridge transport
// adapters. The root package owns the aggregate services contract; subpackages
// own concrete capability contracts and adapters.
package capability

import (
	"strings"

	"lina-core/pkg/plugin/capability/contract"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// Services exposes host-owned capabilities that ordinary plugins may consume
// through a stable service set. Implementations are runtime-owned and may return
// nil for plugin-scoped read capabilities until ServicesForPlugin binds a
// plugin identity. Services methods must not expose database query builders,
// HTTP request objects, write seams, or host-internal governance services.
type Services interface {
	// APIDoc returns the API-documentation localization service.
	APIDoc() contract.APIDocService
	// Auth returns the tenant-auth token handoff service.
	Auth() contract.AuthService
	// BizCtx returns the current request business-context projection service.
	BizCtx() contract.BizCtxService
	// Cache returns the plugin-scoped runtime cache service.
	Cache() contract.CacheService
	// Config returns the plugin-scoped static config service.
	Config() contract.ConfigService
	// HostConfig returns the public host configuration service.
	HostConfig() contract.HostConfigService
	// I18n returns the runtime translation service.
	I18n() contract.I18nService
	// Manifest returns the plugin-scoped manifest resource service.
	Manifest() contract.ManifestService
	// Notify returns the host notification service.
	Notify() contract.NotifyService
	// Org returns the organization capability consumer.
	Org() orgcap.Service
	// PluginLifecycle returns the plugin lifecycle orchestration service.
	PluginLifecycle() contract.PluginLifecycleService
	// PluginSettings returns the host namespaced key-value settings adapter
	// that source plugins use to persist per-plugin configuration in the
	// shared sys_config table instead of owning private SQL schemas.
	PluginSettings() contract.PluginSettingsService
	// PluginState returns the plugin state and enablement service.
	PluginState() contract.PluginStateService
	// Route returns the dynamic-route metadata service.
	Route() contract.RouteService
	// Session returns the online-session service.
	Session() contract.SessionService
	// Tenant returns the tenant capability consumer.
	Tenant() tenantcap.Service
}

// ScopedServicesFactory is implemented by service sets that can return
// a plugin-bound capability view.
type ScopedServicesFactory interface {
	// ForPlugin returns a service set bound to pluginID.
	ForPlugin(pluginID string) Services
}

// ServicesForPlugin returns a plugin-bound capability service set when supported;
// otherwise it returns the supplied services unchanged.
func ServicesForPlugin(services Services, pluginID string) Services {
	if services == nil {
		return nil
	}
	if scoped, ok := services.(ScopedServicesFactory); ok {
		return scoped.ForPlugin(strings.TrimSpace(pluginID))
	}
	return services
}
