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
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/contract"
)

// ConfigureWasmHostServices wires dynamic-plugin host-service dispatchers to
// the same runtime-owned services used by the host HTTP process.
func ConfigureWasmHostServices(
	kvCacheSvc kvcache.Service,
	lockSvc hostlock.Service,
	notifySvc notifysvc.Service,
	configSvc configsvc.PluginConfigReader,
	hostServices capability.Services,
	configFactory contract.ConfigServiceFactory,
	hostConfigSvc contract.HostConfigService,
	manifestFactory contract.ManifestServiceFactory,
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
	if err := wasm.ConfigureAITextHostService(hostServices); err != nil {
		return gerror.Wrap(err, "configure wasm ai text host service failed")
	}
	if err := wasm.ConfigureOrgHostService(hostServices); err != nil {
		return gerror.Wrap(err, "configure wasm org host service failed")
	}
	if err := wasm.ConfigureTenantHostService(hostServices); err != nil {
		return gerror.Wrap(err, "configure wasm tenant host service failed")
	}
	if err := wasm.ConfigureConfigHostService(configFactory); err != nil {
		return gerror.Wrap(err, "configure wasm config host service failed")
	}
	if err := wasm.ConfigureHostConfigService(hostConfigSvc); err != nil {
		return gerror.Wrap(err, "configure wasm host config service failed")
	}
	if err := wasm.ConfigureManifestHostService(manifestFactory); err != nil {
		return gerror.Wrap(err, "configure wasm manifest host service failed")
	}
	return nil
}
