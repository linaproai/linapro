// Package pluginbridge provides the public dynamic-plugin bridge SDK.
//
// Dynamic plugins use this package for declaration-time plugin startup facades,
// route execution helpers, request/response binding, raw host-call transport,
// and guest-side host-service capability clients. Wire protocol contracts, ABI
// constants, artifact metadata, host-call payloads, host-service payloads, and
// codecs are owned by the protocol subpackage.
package pluginbridge

import (
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/dictcap"
	"lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
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
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/plugin/pluginbridge/recordstore"
)

// Services exposes guest-side host-service clients using capability-directory
// semantics. Methods return lightweight clients; zero-value client
// implementations are safe to call but may return transport errors when the
// current process is not running inside a dynamic-plugin WASI guest.
type Services interface {
	// APIDoc returns the API-documentation localization guest client.
	APIDoc() apidoccap.Service
	// Auth returns the authentication and authorization guest namespace.
	Auth() authcap.Service
	// Runtime returns the runtime host service guest client for plugin logs,
	// runtime state, host time, UUID generation, and node identity reads.
	Runtime() RuntimeHostService
	// Storage returns the plugin-scoped object storage domain guest client.
	Storage() storagecap.Service
	// Network returns the governed outbound network host service guest client. Calls
	// are constrained by host-side dynamic-plugin host service authorization.
	Network() NetworkHostService
	// RecordStore returns the governed record store facade for authorized dynamic-plugin table
	// reads and mutations.
	RecordStore() *recordstore.DB
	// Cache returns the plugin-scoped runtime cache domain guest client.
	Cache() cachecap.Service
	// Lock returns the plugin-scoped distributed lock domain guest client.
	Lock() lockcap.Service
	// Plugins returns the plugin-domain guest capability namespace.
	Plugins() plugincap.Service
	// HostConfig returns the authorized host configuration capability client.
	HostConfig() hostconfigcap.Service
	// Manifest returns the plugin-scoped manifest-resource capability client.
	Manifest() manifestcap.Service
	// Users returns the user-domain capability guest client. The returned client
	// exposes visible projections, candidate search, and visibility checks only.
	Users() usercap.Service
	// BizCtx returns the current request business-context guest client.
	BizCtx() bizctxcap.Service
	// Dict returns the dictionary-domain guest client.
	Dict() dictcap.Service
	// Files returns the file-domain guest client.
	Files() filecap.Service
	// Jobs returns the scheduled-job domain guest client.
	Jobs() jobcap.Service
	// Notifications returns the notification-domain ordinary read guest client.
	Notifications() notifycap.Service
	// Org returns the organization capability guest client. The returned client
	// never exposes provider internals and safely degrades through Status or
	// Available when host transport fails.
	Org() orgcap.Service
	// Route returns the current dynamic-route metadata guest client.
	Route() routecap.Service
	// Sessions returns the online-session domain guest client.
	Sessions() sessioncap.Service
	// Tenant returns the tenant capability guest client. The returned client
	// never exposes provider internals and safely degrades through Status or
	// Available when host transport fails.
	Tenant() tenantcap.Service
}

// Declarations exposes declaration-time capabilities for dynamic plugin startup
// code. These methods are for build-time route declarations and host-driven
// discovery executions; runtime business logic should use Services for
// ordinary domain capabilities.
type Declarations interface {
	// Routes returns the dynamic route declaration facade.
	Routes() RouteDeclarations
	// Jobs returns the built-in Jobs declaration facade.
	Jobs() JobDeclarations
}

// RouteDeclarations records build-time route group bindings for dynamic
// plugins. Implementations are owned by the dynamic plugin builder or a no-op
// runtime declaration facade.
type RouteDeclarations interface {
	// Group binds one backend/api-relative package path to a plugin-owned route
	// prefix. The apiPackage value uses slash-separated paths such as
	// "dynamic/v1" and never includes the generated backend/api directory.
	Group(prefix string, apiPackage string) error
}

// JobDeclarations records built-in Jobs declarations during host-driven dynamic
// Jobs discovery.
type JobDeclarations interface {
	// Register submits one dynamic-plugin job declaration to the current
	// host-side Jobs discovery collector.
	Register(contract *protocol.JobContract) error
}

// DeclarationOption customizes one dynamic plugin declaration facade.
type DeclarationOption func(*declarations)

// New creates a guest-side capability directory backed by pluginbridge transport.
func New() Services {
	return directory{}
}

// Default returns the process-default guest-side capability directory.
func Default() Services {
	return defaultDirectory
}
