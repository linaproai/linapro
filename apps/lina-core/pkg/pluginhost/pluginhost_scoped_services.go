// This file provides a plugin-scoped host service directory wrapper so source
// plugin registration and callback flows can bind plugin identity without
// depending on host-internal service packages.

package pluginhost

import (
	"strings"
)

// ScopedHostServicesFactory is implemented by host service directories that
// can return a plugin-bound service view.
type ScopedHostServicesFactory interface {
	// ForPlugin returns a host service directory bound to pluginID.
	ForPlugin(pluginID string) HostServices
}

// HostServicesForPlugin returns a plugin-bound host service directory when
// the base directory supports scoped views. Otherwise it returns the base
// directory unchanged for backward-compatible tests that do not use cache.
func HostServicesForPlugin(services HostServices, pluginID string) HostServices {
	if services == nil {
		return nil
	}
	if scoped, ok := services.(ScopedHostServicesFactory); ok {
		return scoped.ForPlugin(strings.TrimSpace(pluginID))
	}
	return services
}
