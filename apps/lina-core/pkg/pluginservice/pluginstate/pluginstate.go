// Package pluginstate exposes source-plugin enablement lookups without
// publishing host-internal plugin service packages to plugin implementations.
package pluginstate

import (
	"context"
	"strings"

	internalplugin "lina-core/internal/service/plugin"
)

// Service defines plugin enablement lookups published to source plugins.
type Service interface {
	// IsEnabled reports whether the plugin is currently installed and enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
}

// hostPluginEnablement defines the host capability used by the pluginstate adapter.
type hostPluginEnablement interface {
	// IsEnabled reports whether the plugin is currently installed and enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
}

// serviceAdapter bridges host plugin state into the published plugin contract.
type serviceAdapter struct {
	service hostPluginEnablement
}

// New creates and returns the published plugin state service adapter.
func New() Service {
	return &serviceAdapter{service: internalplugin.New(nil)}
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
