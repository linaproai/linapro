// Package pluginstate exposes source-plugin enablement lookups without
// publishing host-internal plugin service packages to plugin implementations.
package pluginstate

import (
	"context"
	"strings"

	"lina-core/pkg/pluginservice/contract"
)

// serviceAdapter bridges host plugin state into the published plugin contract.
type serviceAdapter struct {
	service contract.EnablementReader
}

// New creates and returns a plugin state service backed by the given reader.
func New(service contract.EnablementReader) contract.PluginStateService {
	return &serviceAdapter{service: service}
}

// IsEnabled reports whether the plugin is currently installed and enabled.
func (s *serviceAdapter) IsEnabled(ctx context.Context, pluginID string) bool {
	if s == nil || s.service == nil {
		return false
	}
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return false
	}
	return s.service.IsEnabled(ctx, normalizedPluginID)
}
