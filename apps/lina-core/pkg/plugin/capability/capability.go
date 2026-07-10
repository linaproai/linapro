// Package capability defines the unified host capability services exposed to
// source plugins directly and to dynamic plugins through bridge transport
// adapters. The root package owns the aggregate services contract; subpackages
// own concrete capability contracts and adapters.
package capability

import (
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/dictcap"
	"lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/i18ncap"
	"lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/notifycap"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	"lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/usercap"
)

// Services exposes host-owned capabilities that ordinary plugins may consume
// through a stable service set. Implementations are runtime-owned and may
// return nil for plugin-scoped read capabilities until the host binds a plugin
// identity. Services methods must not expose database query builders, HTTP
// request objects, write seams, or ungoverned host-internal services.
type Services interface {
	// APIDoc returns the API-documentation localization service.
	APIDoc() apidoccap.Service
	// Auth returns the authentication and authorization capability namespace.
	Auth() authcap.Service
	// Users returns the user-domain ordinary capability service.
	Users() usercap.Service
	// BizCtx returns the current request business-context service.
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
	// Jobs returns the scheduled-job domain ordinary capability service.
	Jobs() jobcap.Service
	// Lock returns the plugin-scoped distributed lock service.
	Lock() lockcap.Service
	// Manifest returns the plugin-scoped manifest resource service.
	Manifest() manifestcap.Service
	// Notifications returns the notification-domain ordinary capability service.
	Notifications() notifycap.Service
	// Org returns the organization capability consumer.
	Org() orgcap.Service
	// Plugins returns the governed plugin-domain capability service.
	Plugins() plugincap.Service
	// Route returns the dynamic-route metadata service.
	Route() routecap.Service
	// Sessions returns the online-session domain ordinary capability service.
	Sessions() sessioncap.Service
	// Storage returns the plugin-scoped object storage service.
	Storage() storagecap.Service
	// Tenant returns the tenant capability consumer.
	Tenant() tenantcap.Service
}
