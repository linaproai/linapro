// Package guest exposes the dynamic-plugin side of capability.Services.
// It keeps plugin-facing capability-directory semantics in capability while
// delegating transport details to pluginbridge/guest.
package guest

import (
	plugindata "lina-core/pkg/plugin/capability/data"
)

// Services exposes guest-side host-service clients using capability-directory
// semantics. Methods return lightweight clients; zero-value client
// implementations are safe to call but may return transport errors when the
// current process is not running inside a dynamic-plugin WASI guest.
type Services interface {
	// Runtime returns the runtime host service guest client for plugin logs,
	// runtime state, host time, UUID generation, and node identity reads.
	Runtime() RuntimeHostService
	// Storage returns the governed storage host service guest client for
	// plugin-scoped object reads, writes, deletion, listing, and metadata reads.
	Storage() StorageHostService
	// Network returns the governed outbound network host service guest client. Calls
	// are constrained by host-side dynamic-plugin host service authorization.
	Network() NetworkHostService
	// Data returns the governed data facade for authorized dynamic-plugin table
	// reads and mutations.
	Data() *plugindata.DB
	// Cache returns the governed cache host service guest client for
	// plugin-authorized namespaces and keys.
	Cache() CacheHostService
	// Lock returns the governed distributed lock host service guest client.
	Lock() LockHostService
	// Config returns the plugin-scoped static config host service guest client.
	Config() ConfigHostService
	// Notify returns the governed notification host service guest client.
	Notify() NotifyHostService
	// Cron returns the cron declaration host service guest client used during
	// dynamic-plugin discovery and registration.
	Cron() CronHostService
	// HostConfig returns the authorized host config guest client.
	HostConfig() HostConfigHostService
	// Manifest returns the plugin-scoped manifest-resource guest client.
	Manifest() ManifestHostService
	// Org returns the organization capability guest client. The returned client
	// never exposes provider internals and reports unavailable capability state
	// through Status or Available errors when host transport fails.
	Org() OrgService
	// Tenant returns the tenant capability guest client. The returned client
	// never exposes provider internals and reports unavailable capability state
	// through Status or Available errors when host transport fails.
	Tenant() TenantService
	// AI returns the guest-side AI capability namespace.
	AI() AIService
}

// New creates a guest-side capability directory backed by pluginbridge
// transport.
func New() Services {
	return directory{}
}

// Default returns the process-default guest-side capability directory.
func Default() Services {
	return defaultDirectory
}
