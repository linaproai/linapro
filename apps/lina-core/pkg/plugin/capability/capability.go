// Package capability defines the unified host capability services exposed to
// source plugins directly and to dynamic plugins through bridge transport
// adapters. The root package owns the aggregate services contract; subpackages
// own concrete capability contracts and adapters.
package capability

import (
	"strings"

	"lina-core/pkg/plugin/capability/aicap"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/configcap"
	"lina-core/pkg/plugin/capability/dictcap"
	"lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/i18ncap"
	"lina-core/pkg/plugin/capability/infracap"
	"lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/notifycap"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	"lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/usercap"
)

// Services exposes host-owned capabilities that ordinary plugins may consume
// through a stable service set. Implementations are runtime-owned and may return
// nil for plugin-scoped read capabilities until ServicesForPlugin binds a
// plugin identity. Services methods must not expose database query builders,
// HTTP request objects, write seams, or host-internal governance services.
type Services interface {
	// APIDoc returns the API-documentation localization service.
	APIDoc() apidoccap.Service
	// Auth returns the authentication and authorization capability namespace.
	Auth() authcap.Service
	// AI returns the host AI capability namespace.
	AI() aicap.Service
	// Users returns the user-domain ordinary capability service.
	Users() usercap.Service
	// BizCtx returns the current request business-context projection service.
	BizCtx() bizctxcap.Service
	// Cache returns the plugin-scoped runtime cache service.
	Cache() cachecap.Service
	// Dict returns the dictionary-domain ordinary capability service.
	Dict() dictcap.Service
	// Files returns the file-domain ordinary capability service.
	Files() filecap.Service
	// HostConfig returns the host configuration service.
	HostConfig() hostconfigcap.Service
	// I18n returns the runtime translation service.
	I18n() i18ncap.Service
	// Infra returns the infrastructure-domain ordinary capability service.
	Infra() infracap.Service
	// Jobs returns the scheduled-job domain ordinary capability service.
	Jobs() jobcap.Service
	// Manifest returns the plugin-scoped manifest resource service.
	Manifest() manifestcap.Service
	// Notifications returns the notification-domain ordinary capability service.
	Notifications() notifycap.Service
	// Org returns the organization capability consumer.
	Org() orgcap.Service
	// Plugins returns the plugin-governance ordinary capability service.
	Plugins() plugincap.Service
	// Route returns the dynamic-route metadata service.
	Route() routecap.Service
	// Sessions returns the online-session domain ordinary capability service.
	Sessions() sessioncap.Service
	// Tenant returns the tenant capability consumer.
	Tenant() tenantcap.Service
}

// AdminServices exposes host-owned domain management commands for trusted
// source plugins. The directory is intentionally separate from ordinary
// Services so source plugins can request management capabilities explicitly
// through pluginhost.Services. Methods return narrow domain AdminService
// contracts; callers should inject those narrow interfaces into business
// services instead of storing the whole directory.
type AdminServices interface {
	// Users returns user-domain management commands.
	Users() usercap.AdminService
	// Auth returns authentication and authorization management commands.
	Auth() authcap.AdminService
	// Dict returns dictionary-domain management commands.
	Dict() dictcap.AdminService
	// Files returns file-domain management commands.
	Files() filecap.AdminService
	// Sessions returns online-session management commands.
	Sessions() sessioncap.AdminService
	// Config returns runtime configuration management commands.
	Config() configcap.AdminService
	// Notifications returns notification management commands.
	Notifications() notifycap.AdminService
	// Plugins returns plugin-governance management commands.
	Plugins() plugincap.AdminService
	// Jobs returns scheduled-job management commands.
	Jobs() jobcap.AdminService
	// Infra returns infrastructure management commands.
	Infra() infracap.AdminService
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
