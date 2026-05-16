// This file exposes explicit runtime wiring for dynamic-plugin Wasm host
// services without leaking internal wasm packages to HTTP startup code.

package plugin

import (
	"github.com/gogf/gf/v2/errors/gerror"

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
) error {
	if err := wasm.ConfigureCacheHostService(kvCacheSvc); err != nil {
		return gerror.Wrap(err, "configure wasm cache host service failed")
	}
	if err := wasm.ConfigureLockHostService(lockSvc); err != nil {
		return gerror.Wrap(err, "configure wasm lock host service failed")
	}
	if err := wasm.ConfigureNotifyHostService(notifySvc); err != nil {
		return gerror.Wrap(err, "configure wasm notify host service failed")
	}
	if err := wasm.ConfigureStorageHostService(configSvc); err != nil {
		return gerror.Wrap(err, "configure wasm storage host service failed")
	}
	if err := wasm.ConfigureConfigHostService(configAdapter); err != nil {
		return gerror.Wrap(err, "configure wasm config host service failed")
	}
	return nil
}
