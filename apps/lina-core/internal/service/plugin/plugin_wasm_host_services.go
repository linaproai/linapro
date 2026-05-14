// This file exposes explicit runtime wiring for dynamic-plugin Wasm host
// services without leaking internal wasm packages to HTTP startup code.

package plugin

import (
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/hostlock"
	"lina-core/internal/service/kvcache"
	notifysvc "lina-core/internal/service/notify"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/pkg/pluginservice/contract"
)

// ConfigureWasmHostServices wires dynamic-plugin host-service dispatchers to
// the same runtime-owned services used by the host HTTP process.
func ConfigureWasmHostServices(
	kvCacheSvc kvcache.Service,
	lockSvc hostlock.Service,
	notifySvc notifysvc.Service,
	configSvc configsvc.PluginConfigReader,
	configAdapter contract.ConfigService,
) {
	wasm.ConfigureCacheHostService(kvCacheSvc)
	wasm.ConfigureLockHostService(lockSvc)
	wasm.ConfigureNotifyHostService(notifySvc)
	wasm.ConfigureStorageHostService(configSvc)
	wasm.ConfigureConfigHostService(configAdapter)
}
