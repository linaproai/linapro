// This file provides guest host-service client factories for the Services
// directory while delegating concrete implementations to internal/domainhostcall.
// The helpers stay unexported so dynamic plugins consume capabilities through
// pluginbridge.Default()/New() instead of root package client facades.

package pluginbridge

import (
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/pluginbridge/internal/domainhostcall"
)

func runtimeCapability() RuntimeHostService {
	return domainhostcall.Runtime(invokeGuestHostService)
}

func storageCapability() storagecap.Service {
	return domainhostcall.Storage(invokeGuestHostService)
}

func networkCapability() NetworkHostService {
	return domainhostcall.Network(invokeGuestHostService)
}

func cacheCapability() cachecap.Service {
	return domainhostcall.Cache(invokeGuestHostService)
}

func lockCapability() lockcap.Service {
	return domainhostcall.Lock(invokeGuestHostService)
}

func hostConfigCapability() hostconfigcap.Service {
	return domainhostcall.HostConfigCapability(invokeGuestHostService)
}

func manifestCapability() manifestcap.Service {
	return domainhostcall.ManifestCapability(invokeGuestHostService)
}
