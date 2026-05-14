// This file defines the host-published service directory exposed to source
// plugin registration callbacks.

package pluginhost

import (
	"lina-core/pkg/pluginservice/contract"
)

// HostServices exposes host-owned pluginservice adapters to source plugins.
type HostServices interface {
	// APIDoc returns the host API-documentation localization adapter.
	APIDoc() contract.APIDocService
	// Auth returns the host tenant-auth adapter.
	Auth() contract.AuthService
	// BizCtx returns the host business-context adapter.
	BizCtx() contract.BizCtxService
	// Config returns the host static configuration adapter.
	Config() contract.ConfigService
	// I18n returns the host runtime translation adapter.
	I18n() contract.I18nService
	// Notify returns the host notification adapter.
	Notify() contract.NotifyService
	// PluginState returns the host plugin enablement adapter.
	PluginState() contract.PluginStateService
	// Route returns the host dynamic-route metadata adapter.
	Route() contract.RouteService
	// Session returns the host online-session adapter.
	Session() contract.SessionService
	// TenantFilter returns the host tenant-filter adapter.
	TenantFilter() contract.TenantFilterService
}
