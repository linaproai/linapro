// This file defines the source-plugin visible plugin-state contract.

package contract

import "context"

// PluginStateService defines plugin enablement lookups published to source plugins.
type PluginStateService interface {
	// IsEnabled reports whether the plugin is currently installed and enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
}

// EnablementReader defines the host capability used by plugin-state lookups.
type EnablementReader interface {
	// IsEnabled reports whether the plugin is currently installed and enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
}
