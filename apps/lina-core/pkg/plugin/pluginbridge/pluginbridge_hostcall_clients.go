// This file exposes public guest host-service client helpers while delegating
// the concrete client implementations to internal/domainhostcall. Keeping the
// public facade here preserves the pluginbridge SDK surface without retaining
// per-domain WASI singletons or non-WASI mirror stubs.

package pluginbridge

import (
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/pluginbridge/internal/domainhostcall"
)

// Runtime returns the runtime host service guest client.
func Runtime() RuntimeHostService {
	return runtimeCapability()
}

// Storage returns the storage domain guest client.
func Storage() storagecap.Service {
	return storageCapability()
}

// Network returns the outbound network host service guest client.
func Network() NetworkHostService {
	return networkCapability()
}

// Cache returns the distributed cache domain guest client.
func Cache() cachecap.Service {
	return cacheCapability()
}

// Lock returns the distributed lock domain guest client.
func Lock() lockcap.Service {
	return lockCapability()
}

// HostConfig returns the host config guest client.
func HostConfig() HostConfigHostService {
	return hostConfigClient()
}

// Manifest returns the plugin manifest-resource guest client.
func Manifest() ManifestHostService {
	return manifestClient()
}

// HostLog writes one runtime log entry through the host.
func HostLog(level int, message string, fields map[string]string) error {
	return Runtime().Log(level, message, fields)
}

// HostStateGet reads one plugin-scoped runtime state value.
func HostStateGet(key string) (string, bool, error) {
	return Runtime().StateGet(key)
}

// HostStateGetMany reads plugin-scoped runtime state values.
func HostStateGetMany(keys []string) (map[string]string, []string, error) {
	return Runtime().StateGetMany(keys)
}

// HostStateSet writes one plugin-scoped runtime state value.
func HostStateSet(key string, value string) error {
	return Runtime().StateSet(key, value)
}

// HostStateSetMany writes plugin-scoped runtime state values.
func HostStateSetMany(values map[string]string) error {
	return Runtime().StateSetMany(values)
}

// HostStateDelete removes one plugin-scoped runtime state value.
func HostStateDelete(key string) error {
	return Runtime().StateDelete(key)
}

// HostStateDeleteMany removes plugin-scoped runtime state values.
func HostStateDeleteMany(keys []string) error {
	return Runtime().StateDeleteMany(keys)
}

// HostStateGetInt reads one integer plugin-scoped runtime state value.
func HostStateGetInt(key string) (int, bool, error) {
	return Runtime().StateGetInt(key)
}

// HostStateSetInt writes one integer runtime state value.
func HostStateSetInt(key string, value int) error {
	return Runtime().StateSetInt(key, value)
}

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

func hostConfigClient() HostConfigHostService {
	return domainhostcall.HostConfig(invokeGuestHostService)
}

func hostConfigCapability() hostconfigcap.Service {
	return domainhostcall.HostConfigCapability(invokeGuestHostService)
}

func manifestClient() ManifestHostService {
	return domainhostcall.Manifest(invokeGuestHostService)
}

func manifestCapability() manifestcap.Service {
	return domainhostcall.ManifestCapability(invokeGuestHostService)
}
