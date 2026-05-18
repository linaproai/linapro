// This file contains source-plugin enablement lookup behavior. It keeps reader
// nil handling and plugin ID normalization outside the package entrypoint while
// preserving the published plugin state service contract.

package pluginstate

import (
	"context"
	"strings"
)

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
